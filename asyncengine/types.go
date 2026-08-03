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
	// Meta 提交时绑定的自定义元数据，原样回传。
	Meta map[string]string
}

// Config 创建 Runner 时的配置项。
type Config struct {
	// Namespace 服务命名空间，必填，用于隔离 Redis key（如 "api-pet"）。
	// 同一服务的多实例应使用相同值；不同服务应使用不同值，避免 pending 对账互相影响。
	// 键名形如：asyncengine:{Namespace}:result:{taskID}、asyncengine:{Namespace}:pending。
	Namespace string
	// Workers 并发 worker 数量；<=0 时使用默认值 4。
	Workers int
	// TaskTimeout 单个任务最长执行时间；<=0 时默认 10 分钟。
	// 超时后 ctx 取消，任务函数应感知并尽快返回。
	// 若 TaskFn 返回 error，按该错误标 FAILED；
	// 若 TaskFn 在超时后仍返回 nil，引擎会按 "task timeout" 标 FAILED。
	TaskTimeout time.Duration
	// OnComplete 任务终态回调（成功或运行中失败）；可为 nil。
	// 进程重启对账导致的失败不会触发该回调（notifyCB=false）。
	OnComplete func(Completion)
}
