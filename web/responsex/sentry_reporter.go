package responsex

import (
	"context"
	"fmt"
	"strconv"

	"github.com/getsentry/sentry-go"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/betacats/go-core/utils/envx"
)

const defaultSentryReportTitle = "Request Failed"

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

	code := payload.Parsed.Code
	if code == 0 {
		code = payload.Response.Code
	}
	msg := payload.Parsed.Msg
	if msg == "" {
		msg = payload.Response.Msg
	}

	err := buildSentryErrorMessage(defaultSentryReportTitle, meta.Method, meta.Path, traceID, env, code, msg, meta.Body)

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
		scope.SetTag("errCode", strconv.Itoa(code))
		scope.SetTag("errMsg", msg)

		scope.SetFingerprint([]string{
			sentryFingerprintPart(env),
			sentryFingerprintPart(meta.Path),
			sentryFingerprintPart(meta.Method),
			strconv.Itoa(code),
			sentryFingerprintPart(msg),
		})

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

func buildSentryErrorMessage(title, method, path, traceID, env string, code int, msg, body string) error {
	return fmt.Errorf(
		"[%s]\n"+
			"  Method:   %s %s\n"+
			"  Trace ID: %s\n"+
			"  Env:      %s\n"+
			"  Error:    [%d] %s\n"+
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

