package asyncengine

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// defaultKeyPrefix Redis key 默认前缀。
	defaultKeyPrefix = "asyncengine"
	// defaultWorkers 默认并发 worker 数。
	defaultWorkers = 4
	// defaultTaskTimeout 默认单任务超时时间。
	defaultTaskTimeout = 10 * time.Minute
	// defaultQueueSize 内存任务队列缓冲长度；满时 SubmitFunc 返回 ErrQueueFull。
	defaultQueueSize = 1024
	// defaultResultTTL 任务结果在 Redis 中的过期时间。
	defaultResultTTL = 24 * time.Hour
	// writeTimeout 写 Redis 终态/记录时的独立超时；与任务执行超时分离，
	// 保证任务执行超时后终态仍能正常落库。
	writeTimeout = 5 * time.Second

	// StatePending 任务已提交、尚未被 worker 取走执行。
	StatePending = "PENDING"
	// StateRunning 任务已被 worker 取走、正在执行。
	StateRunning = "RUNNING"
	// StateSucceeded 任务执行成功。
	StateSucceeded = "SUCCEEDED"
	// StateFailed 任务执行失败（含进程重启对账标记失败）。
	StateFailed = "FAILED"
	// restartFailMsg 进程重启时，对账将未完成 PENDING / RUNNING 任务标为失败时写入的错误文案。
	restartFailMsg = "服务重启，任务已终止，请重新提交"
)

// ErrQueueFull 内存任务队列已满，SubmitFunc 拒绝入队（不阻塞调用方）。
var ErrQueueFull = errors.New("async task queue is full")

// ErrParentCtxRequired SubmitFunc 的 parentCtx 参数不可为 nil。
var ErrParentCtxRequired = errors.New("parentCtx is required")

// TaskRecord Redis 中持久化的任务记录（JSON）。
type TaskRecord struct {
	// State 任务状态：PENDING / RUNNING / SUCCEEDED / FAILED。
	State string `json:"state"`
	// Result 成功时的业务结果；失败或未完成时为空。
	Result *Result `json:"result,omitempty"`
	// Error 失败时的错误信息；成功时为空。
	Error string `json:"error,omitempty"`
	// Notify 提交时绑定的任务归属，供鉴权与通知使用。
	Notify NotifyTarget `json:"notify,omitzero"`
}

// task 内存队列中的待执行单元（不落 Redis，仅进程内传递）。
type task struct {
	// id 任务 ID（UUID）。
	id string
	// parentCtx 提交时的请求上下文，用于 WithoutCancel 保留 trace 等值。
	parentCtx context.Context
	// fn 实际执行的闭包。
	fn TaskFn
	// notify 归属信息，执行结束写回记录 / 回调时使用。
	notify NotifyTarget
}

// Runner 异步任务运行器。
// 调用方通过 SubmitFunc 提交闭包，内部经 taskCh 投递给 worker 执行；
// 状态与结果写入 Redis，可通过 GetRecord 查询。
type Runner struct {
	// rdb Redis 客户端，用于存任务记录与 pending 集合。
	rdb *redis.Client
	// taskCh 内存任务队列；缓冲满时提交失败。
	taskCh chan task
	// timeout 单任务执行超时。
	timeout time.Duration
	// keyPrefix Redis key 前缀。
	keyPrefix string
	// resultTTL 结果 key 的过期时间。
	resultTTL time.Duration
	// onComplete 终态回调；可为 nil。
	onComplete func(Completion)
	// logger 可选日志；可为 nil。
	logger Logger
	// wg 等待所有 worker 退出（Shutdown 时使用）。
	wg sync.WaitGroup
}

// MustNew 根据配置创建 Runner，启动 worker 池，并在启动时做 PENDING / RUNNING 对账。
// 对账失败只打日志，不会导致创建失败。
//
// 入参:
//   - rdb: Redis 客户端，不可为 nil（调用方需保证可用）；
//   - cfg: Workers / TaskTimeout / KeyPrefix / OnComplete / Logger 等配置，
//     数值型字段 <=0 或空字符串时使用包内默认值。
//
// 出参:
//   - *Runner: 已启动 worker 的运行器；调用方应在进程退出前调用 Shutdown。
func MustNew(rdb *redis.Client, cfg Config) *Runner {
	if cfg.Workers <= 0 {
		cfg.Workers = defaultWorkers
	}
	if cfg.TaskTimeout <= 0 {
		cfg.TaskTimeout = defaultTaskTimeout
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = defaultKeyPrefix
	}
	runner := &Runner{
		rdb:        rdb,
		taskCh:     make(chan task, defaultQueueSize),
		timeout:    cfg.TaskTimeout,
		keyPrefix:  cfg.KeyPrefix,
		resultTTL:  defaultResultTTL,
		onComplete: cfg.OnComplete,
		logger:     cfg.Logger,
	}
	if err := runner.reconcileOnStart(context.Background()); err != nil {
		runner.errorf("asyncengine reconcile on start: %v", err)
	}
	for i := 0; i < cfg.Workers; i++ {
		runner.wg.Add(1)
		go runner.worker()
	}
	return runner
}

// Shutdown 关闭任务队列并等待所有 worker 退出。
// 关闭后不应再调用 SubmitFunc；已在队列中的任务会继续被消费完毕。
//
// 入参: 无。
// 出参: 无。
func (runner *Runner) Shutdown() {
	close(runner.taskCh)
	runner.wg.Wait()
}

// SubmitFunc 提交异步任务，立即返回 taskID，不等待执行完成。
//
// 流程: 生成 UUID → 写 Redis PENDING + pending 集合 → 非阻塞投递到 taskCh。
// 若队列已满，会回滚刚写入的 Redis 记录并返回 ErrQueueFull。
//
// 入参:
//   - parentCtx: 提交方的请求上下文，不可为 nil；异步执行时通过 WithoutCancel 保留 trace 等值。
//   - fn: 任务闭包，不可为 nil；
//   - notify: 通知人/归属，EmployeeNo 必填（查询鉴权与 OnComplete 路由）。
//
// 出参:
//   - taskID: 任务唯一 ID，可用于 GetRecord；
//   - err: parentCtx/fn/notify 校验失败、Redis 写入失败，或 ErrQueueFull。
func (runner *Runner) SubmitFunc(parentCtx context.Context, fn TaskFn, notify NotifyTarget) (taskID string, err error) {
	if parentCtx == nil {
		return "", ErrParentCtxRequired
	}
	if fn == nil {
		return "", errors.New("async task is nil")
	}
	if notify.EmployeeNo == "" {
		return "", errors.New("notify employeeNo is required")
	}

	taskID = uuid.NewString()
	if err := runner.setStatePending(parentCtx, taskID, notify); err != nil {
		return "", err
	}

	select {
	case runner.taskCh <- task{id: taskID, parentCtx: parentCtx, fn: fn, notify: notify}:
		return taskID, nil
	default:
		// 原子回滚刚写入的 PENDING 记录：Del + SRem 合并为一条事务管道，
		// 避免中途失败留下脏 pending 记录。写入使用独立超时 ctx，
		// 仅继承 parentCtx 的 value（如 trace），不继承其取消/超时。
		rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), writeTimeout)
		defer cancel()
		pipe := runner.rdb.TxPipeline()
		pipe.Del(rollbackCtx, runner.getWholeKeyByTaskId(taskID))
		pipe.SRem(rollbackCtx, runner.pendingSetKey(), taskID)
		if _, err := pipe.Exec(rollbackCtx); err != nil {
			runner.errorf("asyncengine rollback pending: taskId=%s err=%v", taskID, err)
		}
		return "", ErrQueueFull
	}
}

// GetRecord 返回 Redis 中的完整任务记录（含 State / Result / Error / Notify），
// 供业务侧做鉴权或展示，不做状态语义转换。
//
// 入参:
//   - ctx: Redis 操作上下文；
//   - taskID: 任务 ID。
//
// 出参:
//   - *TaskRecord: 记录存在时非 nil；key 不存在时 (nil, nil)；
//   - error: Redis 或 JSON 反序列化错误。
func (runner *Runner) GetRecord(ctx context.Context, taskID string) (*TaskRecord, error) {
	return runner.getValueByTaskId(ctx, taskID)
}

// reconcileOnStart 进程启动时对账：将 pending 集合中仍为 PENDING / RUNNING 的任务
// 全部标记为 FAILED（闭包无法跨进程恢复），并写入 restartFailMsg。
// 对账失败不会触发 OnComplete。
//
// 入参:
//   - ctx: Redis 操作上下文。
//
// 出参:
//   - error: 读取 pending 集合失败时返回；单个任务处理失败仅打日志。
func (runner *Runner) reconcileOnStart(ctx context.Context) error {
	ids, err := runner.rdb.SMembers(ctx, runner.pendingSetKey()).Result()
	if err != nil {
		return err
	}
	for _, id := range ids {
		taskRecord, err := runner.getValueByTaskId(ctx, id)
		if err != nil {
			runner.errorf("asyncengine reconcile get %s: %v", id, err)
			continue
		}
		if taskRecord == nil || (taskRecord.State != StatePending && taskRecord.State != StateRunning) {
			_ = runner.rdb.SRem(ctx, runner.pendingSetKey(), id).Err()
			continue
		}
		if err := runner.setStateFailed(ctx, id, taskRecord.Notify, errors.New(restartFailMsg), false); err != nil {
			runner.errorf("asyncengine reconcile fail %s: %v", id, err)
			continue
		}
		runner.infof("asyncengine reconcile: taskId=%s marked failed(restart)", id)
	}
	return nil
}

// worker 单个 worker 循环：从 taskCh 取任务并执行，直到 channel 关闭。
//
// 入参: 无（方法接收者为 Runner）。
// 出参: 无；退出前 wg.Done()。
func (runner *Runner) worker() {
	defer runner.wg.Done()
	for t := range runner.taskCh {
		runner.run(t)
	}
}

// run 执行单个任务：带超时调用 TaskFn，按结果写 SUCCEEDED / FAILED，
// 并触发 OnComplete；捕获 panic 后记为 internal panic 失败。
func (runner *Runner) run(t task) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			runner.errorf("asyncengine panic: taskId=%s err=%v", t.id, panicVal)
			// 传 parentCtx 而非 Background，保证 panic 落库时 trace 不断链；
			// 写入使用独立超时 ctx，不依赖可能已取消的任务执行 ctx。
			if failErr := runner.setStateFailed(t.parentCtx, t.id, t.notify, errors.New("internal panic"), true); failErr != nil {
				runner.errorf("asyncengine finish failed after panic: taskId=%s err=%v", t.id, failErr)
			}
		}
	}()

	// 执行前标记 RUNNING,让调用方区分"排队中"与"执行中"。
	// 使用独立写入 ctx,仅继承 parentCtx 的 value(如 trace),不继承取消/超时。
	if runErr := runner.setStateRunning(t.parentCtx, t.id, t.notify); runErr != nil {
		runner.errorf("asyncengine mark running: taskId=%s err=%v", t.id, runErr)
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.parentCtx), runner.timeout)
	defer cancel()

	result, err := t.fn(ctx)
	if err != nil {
		// 任务执行 ctx 可能已超时取消，不能直接用于写 Redis；
		// finish 系列内部会基于 parentCtx 创建独立写入 ctx（不继承取消/超时，保留 trace）。
		if failErr := runner.setStateFailed(t.parentCtx, t.id, t.notify, err, true); failErr != nil {
			runner.errorf("asyncengine finish failed: taskId=%s err=%v", t.id, failErr)
		}
		return
	}
	// 超时但 TaskFn 未感知（返回 nil）：ctx 已到 deadline，任务实际未正常完成，
	// 按失败处理，避免被误标 SUCCEEDED。
	if ctx.Err() != nil {
		timeoutErr := errors.New("task timeout")
		runner.errorf("asyncengine task timeout but fn returned nil: taskId=%s", t.id)
		if failErr := runner.setStateFailed(t.parentCtx, t.id, t.notify, timeoutErr, true); failErr != nil {
			runner.errorf("asyncengine finish failed: taskId=%s err=%v", t.id, failErr)
		}
		return
	}
	if succErr := runner.setStateSucceeded(t.parentCtx, t.id, t.notify, result); succErr != nil {
		runner.errorf("asyncengine finish succeeded: taskId=%s err=%v", t.id, succErr)
	}
}

// setStatePending 将任务以 PENDING 状态写入 Redis，并加入 pending 集合（事务管道）。
// 使用独立超时 ctx：仅继承 parentCtx 的 value（如 trace），不继承其取消/超时。
//
// 入参:
//   - parentCtx: 提交方上下文（仅继承 value）；
//   - taskID: 任务 ID；
//   - notify: 归属信息，一并序列化进记录。
//
// 出参:
//   - error: JSON 序列化或 Redis 管道执行失败。
func (runner *Runner) setStatePending(parentCtx context.Context, taskID string, notify NotifyTarget) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), writeTimeout)
	defer cancel()

	taskRecord := TaskRecord{State: StatePending, Notify: notify}
	b, err := json.Marshal(taskRecord)
	if err != nil {
		return err
	}
	pipe := runner.rdb.TxPipeline()
	pipe.Set(ctx, runner.getWholeKeyByTaskId(taskID), b, runner.resultTTL)
	pipe.SAdd(ctx, runner.pendingSetKey(), taskID)
	_, err = pipe.Exec(ctx)
	return err
}

// setStateRunning 将任务标记为 RUNNING（覆盖 PENDING 记录，保留 Notify）。
// 使用独立超时 ctx：仅继承 parentCtx 的 value（如 trace），不继承其取消/超时。
func (runner *Runner) setStateRunning(parentCtx context.Context, taskID string, notify NotifyTarget) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), writeTimeout)
	defer cancel()
	taskRecord := TaskRecord{State: StateRunning, Notify: notify}
	return runner.saveTaskToRedis(ctx, taskID, taskRecord)
}

// setStateSucceeded 将任务标记为 SUCCEEDED，写入结果，从 pending 集合移除，并触发 OnComplete。
//
// 入参:
//   - ctx: Redis 上下文；
//   - taskID: 任务 ID；
//   - notify: 归属信息；
//   - result: 业务成功结果。
//
// 出参:
//   - error: 写 Redis 失败时返回（仍会尝试打错误日志）。
func (runner *Runner) setStateSucceeded(parentCtx context.Context, taskID string, notify NotifyTarget, result Result) error {
	// 写入使用独立超时 ctx：仅继承 parentCtx 的 value（如 trace），
	// 不继承其取消/超时，确保任务执行超时后终态仍能落库。
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), writeTimeout)
	defer cancel()

	taskRecord := TaskRecord{State: StateSucceeded, Result: &result, Notify: notify}
	if err := runner.saveTaskToRedis(ctx, taskID, taskRecord); err != nil {
		runner.errorf("asyncengine save result: taskId=%s err=%v", taskID, err)
		return err
	}
	_ = runner.rdb.SRem(ctx, runner.pendingSetKey(), taskID).Err()
	runner.emitComplete(Completion{TaskID: taskID, Result: result, Notify: notify})
	return nil
}

// setStateFailed 将任务标记为 FAILED，写入错误信息，从 pending 集合移除；
// 按 notifyCB 决定是否触发 OnComplete。
//
// 入参:
//   - ctx: Redis 上下文；
//   - taskID: 任务 ID；
//   - notify: 归属信息；
//   - fail: 失败原因；
//   - notifyCB: true 时调用 OnComplete（运行中失败）；false 时不回调（如启动对账）。
//
// 出参:
//   - error: 写 Redis 失败时返回。
func (runner *Runner) setStateFailed(parentCtx context.Context, taskID string, notify NotifyTarget, fail error, notifyCB bool) error {
	// 写入使用独立超时 ctx：仅继承 parentCtx 的 value（如 trace），
	// 不继承其取消/超时，确保任务执行超时后终态仍能落库。
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), writeTimeout)
	defer cancel()

	taskRecord := TaskRecord{State: StateFailed, Error: fail.Error(), Notify: notify}
	if err := runner.saveTaskToRedis(ctx, taskID, taskRecord); err != nil {
		runner.errorf("asyncengine save failed result: taskId=%s err=%v", taskID, err)
		return err
	}
	_ = runner.rdb.SRem(ctx, runner.pendingSetKey(), taskID).Err()
	if notifyCB {
		runner.emitComplete(Completion{TaskID: taskID, Err: fail, Notify: notify})
	}
	return nil
}

// emitComplete 安全调用 OnComplete：回调为 nil 则直接返回；回调内 panic 会被捕获并打日志。
//
// 入参:
//   - c: 终态回调参数。
//
// 出参: 无。
func (runner *Runner) emitComplete(c Completion) {
	if runner.onComplete == nil {
		return
	}
	defer func() {
		if panicVal := recover(); panicVal != nil {
			runner.errorf("asyncengine OnComplete panic: taskId=%s err=%v", c.TaskID, panicVal)
		}
	}()
	runner.onComplete(c)
}

// saveTaskToRedis 将完整 TaskRecord 序列化后写入 Redis result key，并设置 TTL。
//
// 入参:
//   - ctx: Redis 上下文；
//   - taskID: 任务 ID；
//   - taskRecord: 要持久化的任务记录。
//
// 出参:
//   - error: JSON 序列化或 Redis Set 失败。
func (runner *Runner) saveTaskToRedis(ctx context.Context, taskID string, taskRecord TaskRecord) error {
	b, err := json.Marshal(taskRecord)
	if err != nil {
		return err
	}
	return runner.rdb.Set(ctx, runner.getWholeKeyByTaskId(taskID), b, runner.resultTTL).Err()
}

// getValueByTaskId 从 Redis 读取并反序列化任务记录。
//
// 入参:
//   - ctx: Redis 上下文；
//   - taskID: 任务 ID。
//
// 出参:
//   - *TaskRecord: key 存在时返回指针；key 不存在时 (nil, nil)；
//   - error: Redis 非 Nil 错误，或 JSON 反序列化失败。
func (runner *Runner) getValueByTaskId(ctx context.Context, taskID string) (*TaskRecord, error) {
	val, err := runner.rdb.Get(ctx, runner.getWholeKeyByTaskId(taskID)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var taskRecord TaskRecord
	if err := json.Unmarshal([]byte(val), &taskRecord); err != nil {
		return nil, err
	}
	return &taskRecord, nil
}

// getWholeKeyByTaskId 生成任务结果 Redis key：{keyPrefix}:result:{taskID}。
//
// 入参:
//   - taskID: 任务 ID。
//
// 出参:
//   - string: Redis key。
func (runner *Runner) getWholeKeyByTaskId(taskID string) string {
	return runner.keyPrefix + ":result:" + taskID
}

// pendingSetKey 生成未终态任务集合的 Redis key：{keyPrefix}:pending。
// 集合中包含 PENDING / RUNNING 的 taskId，供启动对账使用。
//
// 入参: 无。
// 出参:
//   - string: Redis Set key。
func (runner *Runner) pendingSetKey() string {
	return runner.keyPrefix + ":pending"
}

// infof 输出信息日志；logger 为 nil 时无操作。
//
// 入参:
//   - format / args: 同 fmt.Printf。
//
// 出参: 无。
func (runner *Runner) infof(format string, args ...any) {
	if runner.logger != nil {
		runner.logger.Infof(format, args...)
	}
}

// errorf 输出错误日志；logger 为 nil 时无操作。
//
// 入参:
//   - format / args: 同 fmt.Printf。
//
// 出参: 无。
func (runner *Runner) errorf(format string, args ...any) {
	if runner.logger != nil {
		runner.logger.Errorf(format, args...)
	}
}
