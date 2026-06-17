package errorx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/getsentry/sentry-go"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/betacats/go-core/utils/envx"
)

// ReadRequestBody 读取并恢复 request body
// 返回 body 字符串和错误信息
// 注意：会自动恢复 body，以便后续 handler 可以再次读取
func ReadRequestBody(r *http.Request) (string, error) {
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	// 恢复 body 以便后续使用
	r.Body = io.NopCloser(bytes.NewBuffer(buf))
	return string(buf), nil
}

// BuildErrorMessage 构建人类可读的错误信息
// 参数说明：
//   - ctx: 上下文（用于获取 traceId）
//   - title: 错误标题（如 "Request Failed" 或 "Panic Recovered"）
//   - method: HTTP 方法
//   - path: 请求路径
//   - env: 环境名称
//   - errorCode: 错误码或错误类型（如 "500" 或 "PANIC"）
//   - errorMsg: 错误消息
//   - body: 请求体内容
func BuildErrorMessage(ctx context.Context, title, method, path, env, errorCode, errorMsg, body string) error {
	traceId := oteltrace.SpanContextFromContext(ctx).TraceID().String()
	return fmt.Errorf(
		"[%s]\n"+
			"  Method:   %s %s\n"+
			"  Trace ID: %s\n"+
			"  Env:      %s\n"+
			"  Error:    [%s] %s\n"+
			"  Body:     %s",
		title,
		method,
		path,
		traceId,
		env,
		errorCode,
		errorMsg,
		body,
	)
}

// ReportToSentry 统一的 Sentry 错误上报函数
// 参数说明：
//   - r: HTTP 请求对象
//   - title: 错误标题（如 "Request Failed" 或 "Panic Recovered"）
//   - errCode: 错误码（字符串格式）
//   - errMsg: 错误消息
func ReportToSentry(r *http.Request, title, errCode, errMsg string) error {
	// 读取请求 body
	bodyStr, readErr := ReadRequestBody(r)
	if readErr != nil {
		return readErr
	}

	// 构建人类可读的错误信息
	err := BuildErrorMessage(
		r.Context(),
		title,
		r.Method,
		r.URL.Path,
		envx.ENV(),
		errCode,
		errMsg,
		bodyStr,
	)

	traceId := oteltrace.SpanContextFromContext(r.Context()).TraceID().String()
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetRequest(r)
		scope.SetTag("traceId", traceId)
		scope.SetTag("method", r.Method)
		scope.SetTag("errCode", errCode)
		scope.SetTag("errMsg", errMsg)
		// 按环境、路径、方法、错误码、错误信息进行 sentry 错误分组
		scope.SetFingerprint([]string{
			envx.ENV(),
			r.URL.Path,
			r.Method,
			errCode,
			errMsg,
		})
		sentry.CaptureException(err)
	})
	return nil
}
