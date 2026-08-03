// Package asyncengine 提供进程内异步任务引擎：
// 调用方通过 SubmitFunc 提交闭包，由 worker 池异步执行，任务状态与结果落 Redis，
// 支持查询结果、启动对账（重启后未完成任务标失败）以及可选的终态回调。
package asyncengine

import (
	"context"
	"time"
)

// Result 异步任务执行成功后的业务结果。
// 所有业务结果统一放在 Data 中，由调用方自行约定结构与字段含义。
type Result struct {
	// Data 任意结构化业务数据，由调用方自行约定类型与含义。
	Data any `json:"data,omitempty"`
}

// NotifyTarget 任务归属与通知人，SubmitFunc 必传。
// EmployeeNo 为通知人，必填；HospitalNo / ShopNo 用于查询鉴权与通知路由。
type NotifyTarget struct {
	// HospitalNo 医院编号。
	HospitalNo string `json:"hospitalNo,omitempty"`
	// ShopNo 门店编号。
	ShopNo string `json:"shopNo,omitempty"`
	// EmployeeNo 通知人（员工编号），必填。
	EmployeeNo string `json:"employeeNo,omitempty"`
}

// TaskFn 异步任务执行体（闭包）。
//
// 入参:
//   - ctx: 带超时的上下文，超时后应尽快返回；超时时长由 Config.TaskTimeout 决定。
//
// 出参:
//   - Result: 成功时的业务结果；
//   - error: 非 nil 时任务标记为 FAILED，错误信息写入 TaskRecord.Error。
type TaskFn func(ctx context.Context) (Result, error)

// Completion 任务进入终态（成功或失败）时传给 OnComplete 的参数。
type Completion struct {
	// TaskID 任务唯一 ID（SubmitFunc 返回的 UUID）。
	TaskID string
	// Result 成功时的结果；失败时为零值。
	Result Result
	// Err 失败原因；成功时为 nil。
	Err error
	// Notify 提交时绑定的归属信息，便于回调侧路由通知。
	Notify NotifyTarget
}

// Logger 可选日志接口；Config.Logger 为 nil 时引擎不输出日志。
type Logger interface {
	// Infof 输出信息级日志，format/args 语义同 fmt.Printf。
	Infof(format string, args ...any)
	// Errorf 输出错误级日志，format/args 语义同 fmt.Printf。
	Errorf(format string, args ...any)
}

// Config 创建 Runner 时的配置项。
type Config struct {
	// Workers 并发 worker 数量；<=0 时使用默认值 4。
	Workers int
	// TaskTimeout 单个任务最长执行时间；<=0 时默认 10 分钟。
	// 超时后 ctx 取消，任务函数应感知并尽快返回。
	// 若 TaskFn 返回 error，按该错误标 FAILED；
	// 若 TaskFn 在超时后仍返回 nil，引擎会按 "task timeout" 标 FAILED。
	TaskTimeout time.Duration
	// KeyPrefix Redis key 前缀，默认 "asyncengine"。
	// 实际 key 形如：{KeyPrefix}:result:{taskID}、{KeyPrefix}:pending。
	KeyPrefix string
	// OnComplete 任务终态回调（成功或运行中失败）；可为 nil。
	// 进程重启对账导致的失败不会触发该回调（notifyCB=false）。
	OnComplete func(Completion)
	// Logger 可选日志实现；nil 表示静默。
	Logger Logger
}
