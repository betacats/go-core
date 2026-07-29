package responsex

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/getsentry/sentry-go"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/betacats/go-core/utils/envx"
)

// PanicError 包装 panic 信息和堆栈，实现了 ResponseCoder 接口
// 可以被 Builder 自动识别，统一错误响应格式和上报逻辑
type PanicError struct {
	result interface{}
	stack  string
}

// Error 实现 error 接口
func (e *PanicError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("panic: %v", e.result)
}

// ResponseCode 实现 ResponseCoder 接口，返回默认错误码
// 注意：实际响应码由 Builder 的 DefaultErrorCode 决定，这里返回 0 表示使用默认值
func (e *PanicError) ResponseCode() int {
	return 0 // 0 表示使用 Builder 配置的 DefaultErrorCode
}

// ResponseMsg 实现 ResponseCoder 接口，返回 panic 信息
func (e *PanicError) ResponseMsg() string {
	if e == nil {
		return "Internal Server Error"
	}
	return fmt.Sprintf("%v", e.result)
}

// Stack 返回 panic 堆栈信息
func (e *PanicError) Stack() string {
	if e == nil {
		return ""
	}
	return e.stack
}

// Result 返回原始的 panic 内容
func (e *PanicError) Result() interface{} {
	if e == nil {
		return nil
	}
	return e.result
}

// NewPanicError 从 recover() 的结果创建 PanicError，自动捕获堆栈
func NewPanicError(result interface{}) *PanicError {
	if result == nil {
		return nil
	}
	return &PanicError{
		result: result,
		stack:  string(debug.Stack()),
	}
}

// IsPanicError 检查错误是否是 PanicError
func IsPanicError(err error) (*PanicError, bool) {
	if err == nil {
		return nil, false
	}
	var pErr *PanicError
	if errors.As(err, &pErr) {
		return pErr, true
	}
	return nil, false
}

// ReportPanic 直接上报 panic 到 Sentry，用于非 HTTP 场景（如 goroutine）
//
// 使用示例：
//
//	go func() {
//	    defer func() {
//	        if result := recover(); result != nil {
//	            responsex.ReportPanic(context.Background(), result)
//	        }
//	    }()
//	    // 你的业务逻辑...
//	}()
func ReportPanic(ctx context.Context, result interface{}) {
	if result == nil {
		return
	}

	panicErr := NewPanicError(result)
	env := envx.ENV()
	traceID := oteltrace.SpanContextFromContext(ctx).TraceID().String()
	stack := panicErr.Stack()

	errMsg := fmt.Sprintf("%v\n\nStack Trace:\n%s", result, stack)
	err := fmt.Errorf(
		"[%s]\n"+
			"  Trace ID: %s\n"+
			"  Env:      %s\n"+
			"  Error:    [%s] %s",
		panicSentryReportTitle,
		traceID,
		env,
		panicSentryErrCode,
		errMsg,
	)

	sentry.WithScope(func(scope *sentry.Scope) {
		if traceID != "" {
			scope.SetTag("traceId", traceID)
		}
		scope.SetTag("errCode", panicSentryErrCode)
		scope.SetTag("errMsg", fmt.Sprintf("%v", result))
		scope.SetLevel(sentry.LevelFatal)

		scope.SetContext("panic", map[string]any{
			"result": fmt.Sprintf("%v", result),
			"stack":  stack,
		})

		// panic 分组策略
		fingerprint := []string{
			env,
			panicSentryErrCode,
			fmt.Sprintf("%v", result),
		}
		if stack != "" {
			if stackLines := bytes.SplitN([]byte(stack), []byte("\n"), 3); len(stackLines) >= 2 {
				fingerprint = append(fingerprint, string(stackLines[1]))
			}
		}
		scope.SetFingerprint(fingerprint)

		sentry.CaptureException(err)
	})
}
