package asyncengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// defaultKeyPrefix Redis key 固定前缀，不对外配置。
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
	// Meta 提交时绑定的自定义元数据（如业务侧的归属、路由信息），引擎原样存储与回传。
	Meta map[string]string `json:"meta,omitempty"`
}

// task 内存队列中的待执行单元（不落 Redis，仅进程内传递）。
type task struct {
	id        string
	parentCtx context.Context
	fn        TaskFn
	meta      map[string]string
}

// Runner 异步任务运行器。
// 调用方通过 SubmitFunc 提交闭包，内部经 taskCh 投递给 worker 执行；
// 状态与结果写入 Redis，可通过 GetRecord 查询。
type Runner struct {
	rdb        *redis.Client
	namespace  string
	taskCh     chan task
	timeout    time.Duration
	resultTTL  time.Duration
	onComplete func(Completion)
	wg         sync.WaitGroup
}

// MustNew 根据配置创建 Runner，启动 worker 池，并在启动时做 PENDING / RUNNING 对账。
// 对账失败会被忽略（不会导致创建失败）。
//
// 入参:
//   - rdb: Redis 客户端，不可为 nil（调用方需保证可用）；
//   - cfg: Namespace / Workers / TaskTimeout / OnComplete 等配置；
//     Namespace 必填；数值型字段 <=0 时使用包内默认值。
//
// 出参:
//   - *Runner: 已启动 worker 的运行器；调用方应在进程退出前调用 Shutdown。
func MustNew(rdb *redis.Client, cfg Config) *Runner {
	namespace := strings.TrimSpace(cfg.Namespace)
	if namespace == "" {
		panic("asyncengine: namespace is required")
	}
	if strings.Contains(namespace, ":") {
		panic(fmt.Sprintf("asyncengine: namespace %q must not contain ':'", namespace))
	}
	if cfg.Workers <= 0 {
		cfg.Workers = defaultWorkers
	}
	if cfg.TaskTimeout <= 0 {
		cfg.TaskTimeout = defaultTaskTimeout
	}
	runner := &Runner{
		rdb:        rdb,
		namespace:  namespace,
		taskCh:     make(chan task, defaultQueueSize),
		timeout:    cfg.TaskTimeout,
		resultTTL:  defaultResultTTL,
		onComplete: cfg.OnComplete,
	}
	_ = runner.reconcileOnStart(context.Background())
	for i := 0; i < cfg.Workers; i++ {
		runner.wg.Add(1)
		go runner.worker()
	}
	return runner
}

// Shutdown 关闭任务队列并等待所有 worker 退出。
// 关闭后不应再调用 SubmitFunc；已在队列中的任务会继续被消费完毕。
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
//   - meta: 自定义元数据，可为 nil；引擎原样落库并通过 GetRecord / OnComplete 回传。
//
// 出参:
//   - taskID: 任务唯一 ID，可用于 GetRecord；
//   - err: parentCtx/fn 校验失败、Redis 写入失败，或 ErrQueueFull。
func (runner *Runner) SubmitFunc(parentCtx context.Context, fn TaskFn, meta map[string]string) (taskID string, err error) {
	if parentCtx == nil {
		return "", ErrParentCtxRequired
	}
	if fn == nil {
		return "", errors.New("async task is nil")
	}

	taskID = uuid.NewString()
	if err := runner.setStatePending(parentCtx, taskID, meta); err != nil {
		return "", err
	}

	select {
	case runner.taskCh <- task{id: taskID, parentCtx: parentCtx, fn: fn, meta: meta}:
		return taskID, nil
	default:
		// 原子回滚刚写入的 PENDING 记录：Del + SRem 合并为一条事务管道。
		rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), writeTimeout)
		defer cancel()
		pipe := runner.rdb.TxPipeline()
		pipe.Del(rollbackCtx, runner.getWholeKeyByTaskId(taskID))
		pipe.SRem(rollbackCtx, runner.pendingSetKey(), taskID)
		_, _ = pipe.Exec(rollbackCtx)
		return "", ErrQueueFull
	}
}

// GetRecord 返回 Redis 中的完整任务记录（含 State / Result / Error / Meta），
// 供业务侧做鉴权或展示，不做状态语义转换。
func (runner *Runner) GetRecord(ctx context.Context, taskID string) (*TaskRecord, error) {
	return runner.getValueByTaskId(ctx, taskID)
}

// reconcileOnStart 进程启动时对账：将 pending 集合中仍为 PENDING / RUNNING 的任务
// 全部标记为 FAILED（闭包无法跨进程恢复），并写入 restartFailMsg。
// 对账失败不会触发 OnComplete；单个任务处理失败会跳过。
func (runner *Runner) reconcileOnStart(ctx context.Context) error {
	ids, err := runner.rdb.SMembers(ctx, runner.pendingSetKey()).Result()
	if err != nil {
		return err
	}
	for _, id := range ids {
		taskRecord, err := runner.getValueByTaskId(ctx, id)
		if err != nil {
			continue
		}
		if taskRecord == nil || (taskRecord.State != StatePending && taskRecord.State != StateRunning) {
			_ = runner.rdb.SRem(ctx, runner.pendingSetKey(), id).Err()
			continue
		}
		_ = runner.setStateFailed(ctx, id, taskRecord.Meta, errors.New(restartFailMsg), false)
	}
	return nil
}

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
		if recover() != nil {
			_ = runner.setStateFailed(t.parentCtx, t.id, t.meta, errors.New("internal panic"), true)
		}
	}()

	_ = runner.setStateRunning(t.parentCtx, t.id, t.meta)

	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.parentCtx), runner.timeout)
	defer cancel()

	result, err := t.fn(ctx)
	if err != nil {
		_ = runner.setStateFailed(t.parentCtx, t.id, t.meta, err, true)
		return
	}
	// 超时但 TaskFn 未感知（返回 nil）：按失败处理，避免被误标 SUCCEEDED。
	if ctx.Err() != nil {
		_ = runner.setStateFailed(t.parentCtx, t.id, t.meta, errors.New("task timeout"), true)
		return
	}
	// 成功落库失败时降级为 FAILED，避免任务一直停在 PENDING/RUNNING。
	// Redis 彻底不可用时两边都写不进去，视为基础设施问题，不再处理。
	if err := runner.setStateSucceeded(t.parentCtx, t.id, t.meta, result); err != nil {
		_ = runner.setStateFailed(t.parentCtx, t.id, t.meta, errors.New("保存任务结果失败"), true)
	}
}

func (runner *Runner) setStatePending(parentCtx context.Context, taskID string, meta map[string]string) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), writeTimeout)
	defer cancel()

	taskRecord := TaskRecord{State: StatePending, Meta: meta}
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

func (runner *Runner) setStateRunning(parentCtx context.Context, taskID string, meta map[string]string) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), writeTimeout)
	defer cancel()
	return runner.saveTaskToRedis(ctx, taskID, TaskRecord{State: StateRunning, Meta: meta})
}

func (runner *Runner) setStateSucceeded(parentCtx context.Context, taskID string, meta map[string]string, result Result) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), writeTimeout)
	defer cancel()

	if err := runner.saveTaskToRedis(ctx, taskID, TaskRecord{State: StateSucceeded, Result: &result, Meta: meta}); err != nil {
		return err
	}
	_ = runner.rdb.SRem(ctx, runner.pendingSetKey(), taskID).Err()
	runner.emitComplete(Completion{TaskID: taskID, Result: result, Meta: meta})
	return nil
}

func (runner *Runner) setStateFailed(parentCtx context.Context, taskID string, meta map[string]string, fail error, notifyCB bool) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), writeTimeout)
	defer cancel()

	if err := runner.saveTaskToRedis(ctx, taskID, TaskRecord{State: StateFailed, Error: fail.Error(), Meta: meta}); err != nil {
		return err
	}
	_ = runner.rdb.SRem(ctx, runner.pendingSetKey(), taskID).Err()
	if notifyCB {
		runner.emitComplete(Completion{TaskID: taskID, Err: fail, Meta: meta})
	}
	return nil
}

func (runner *Runner) emitComplete(c Completion) {
	if runner.onComplete == nil {
		return
	}
	defer func() { _ = recover() }()
	runner.onComplete(c)
}

func (runner *Runner) saveTaskToRedis(ctx context.Context, taskID string, taskRecord TaskRecord) error {
	b, err := json.Marshal(taskRecord)
	if err != nil {
		return err
	}
	return runner.rdb.Set(ctx, runner.getWholeKeyByTaskId(taskID), b, runner.resultTTL).Err()
}

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

func (runner *Runner) getWholeKeyByTaskId(taskID string) string {
	return defaultKeyPrefix + ":" + runner.namespace + ":result:" + taskID
}

func (runner *Runner) pendingSetKey() string {
	return defaultKeyPrefix + ":" + runner.namespace + ":pending"
}
