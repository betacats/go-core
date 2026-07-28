package responsex

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/getsentry/sentry-go"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/betacats/go-core/utils/envx"
)

const (
	defaultSentryReportTitle = "Request Failed"
	panicSentryReportTitle   = "Panic Recovered"
	panicSentryErrCode       = "PANIC"
)

// SentryReporter 是 responsex 内置的默认错误上报实现。
// 当 Builder 开启 EnableReport 且未显式指定 Reporter 时，会自动使用它。
type SentryReporter struct{}

// NewSentryReporter 创建 Sentry 上报器。
func NewSentryReporter() *SentryReporter {
	return &SentryReporter{}
}

// Report 将 ParsedError 和请求元数据上报到 Sentry。
func (r *SentryReporter) Report(ctx context.Context, payload ReporterPayload) {
	if payload.Parsed.Err == nil {
		return
	}

	meta, _ := GetRequestMeta(ctx)
	traceID := oteltrace.SpanContextFromContext(ctx).TraceID().String()

	env := meta.Env
	if env == "" {
		env = envx.ENV()
	}

	// 检查是否是 panic 错误
	panicErr, isPanic := IsPanicError(payload.Parsed.Err)

	title := defaultSentryReportTitle
	code := payload.Parsed.Code
	if code == 0 {
		code = payload.Response.Code
	}
	msg := payload.Parsed.Msg
	if msg == "" {
		msg = payload.Response.Msg
	}

	// panic 特殊处理
	var errCodeTag string
	var stack string
	sentryLevel := sentry.LevelError
	fingerprintParts := []string{
		sentryFingerprintPart(env),
		sentryFingerprintPart(meta.Path),
		sentryFingerprintPart(meta.Method),
	}

	if isPanic {
		title = panicSentryReportTitle
		errCodeTag = panicSentryErrCode
		stack = panicErr.Stack()
		msg = fmt.Sprintf("%v\n\nStack Trace:\n%s", panicErr.Result(), stack)
		sentryLevel = sentry.LevelFatal
		// panic 分组策略：不按 path 分组，同一个 panic 在多个接口触发应该聚合
		fingerprintParts = []string{
			sentryFingerprintPart(env),
			panicSentryErrCode,
			sentryFingerprintPart(fmt.Sprintf("%v", panicErr.Result())),
		}
		// 提取堆栈第一行作为分组依据
		if stack != "" {
			if stackLines := bytes.SplitN([]byte(stack), []byte("\n"), 3); len(stackLines) >= 2 {
				fingerprintParts = append(fingerprintParts, string(stackLines[1]))
			}
		}
	} else {
		errCodeTag = strconv.Itoa(code)
		fingerprintParts = append(fingerprintParts,
			strconv.Itoa(code),
			sentryFingerprintPart(msg),
		)
	}

	err := buildSentryErrorMessage(title, meta.Method, meta.Path, traceID, env, errCodeTag, msg, meta.Body)

	sentry.WithScope(func(scope *sentry.Scope) {
		if traceID != "" {
			scope.SetTag("traceId", traceID)
		}
		if meta.Method != "" {
			scope.SetTag("method", meta.Method)
		}
		if meta.Path != "" {
			scope.SetTag("path", meta.Path)
		}
		scope.SetTag("errCode", errCodeTag)
		scope.SetTag("errMsg", payload.Parsed.Msg)
		scope.SetLevel(sentryLevel)

		// panic 额外设置堆栈上下文
		if isPanic {
			scope.SetContext("panic", map[string]any{
				"result": fmt.Sprintf("%v", panicErr.Result()),
				"stack":  stack,
			})
		}

		scope.SetFingerprint(fingerprintParts)

		scope.SetContext("responsex", map[string]any{
			"result": payload.Response.Result,
			"code":   payload.Response.Code,
			"msg":    payload.Response.Msg,
		})
		if meta.Body != "" {
			scope.SetContext("requestMeta", map[string]any{"body": meta.Body})
		}

		sentry.CaptureException(err)
	})
}

func buildSentryErrorMessage(title, method, path, traceID, env string, code interface{}, msg, body string) error {
	return fmt.Errorf(
		"[%s]\n"+
			"  Method:   %s %s\n"+
			"  Trace ID: %s\n"+
			"  Env:      %s\n"+
			"  Error:    [%v] %s\n"+
			"  Body:     %s",
		title,
		method,
		path,
		traceID,
		env,
		code,
		msg,
		body,
	)
}

func sentryFingerprintPart(v string) string {
	if v == "" {
		return "-"
	}
	return v
}
