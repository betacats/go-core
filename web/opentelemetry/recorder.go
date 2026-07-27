package opentelemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/betacats/go-core/web/errorx"
	"github.com/betacats/go-core/web/responsex"
)

// BodyRecorder 包装 http.ResponseWriter，拦截写入的响应体。
// 用于在 HTTP handler 执行后检查响应内容是否为业务错误。
type BodyRecorder struct {
	// ResponseWriter 原始 HTTP 响应写入器。
	http.ResponseWriter
	// buf 拦截写入数据的缓冲区。
	buf bytes.Buffer
}

// Write 将数据同时写入缓冲区和原始的 ResponseWriter。
func (r *BodyRecorder) Write(data []byte) (int, error) {
	r.buf.Write(data)
	return r.ResponseWriter.Write(data)
}

// Body 返回已拦截的响应体字节切片。
func (r *BodyRecorder) Body() []byte {
	return r.buf.Bytes()
}

// IsBusinessError 检测统一响应体中是否包含业务错误。
// 规则：
//   - "result" 字段为 false
//   - "code" 字段不为 0
func IsBusinessError(body []byte) bool {
	if len(body) == 0 {
		return false
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return false
	}
	if v, ok := m["result"]; ok {
		if b, ok := v.(bool); ok && !b {
			return true
		}
	}
	if v, ok := m["code"]; ok {
		if f, ok := v.(float64); ok && f != 0 {
			return true
		}
	}
	return false
}

// Middleware 返回一个标准 HTTP 中间件，在 handler 执行后检查响应体。
//
// 如果响应体是 go-core 统一格式的业务错误（result=false 或 code!=0），
// 会将当前 span 标记为 codes.Error，确保整条 trace 被全量保留。
//
// 在 go-zero REST 服务中的使用方式：
//
//	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
//		return opentelemetry.Middleware(next).ServeHTTP
//	})
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &BodyRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)

		if IsBusinessError(rec.Body()) {
			if span := trace.SpanFromContext(r.Context()); span.IsRecording() {
				span.SetStatus(codes.Error, "business error")
			}
		}
	})
}

// RecoverMiddleware 返回一个 panic 恢复中间件，捕获 handler 中的 panic。
//
// 它会将当前 span 标记为 codes.Error，确保整条 trace 被全量保留，
// 并向客户端返回统一错误响应。
//
// 在 go-zero REST 服务中应注册为第一个中间件：
//
//	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
//	    return opentelemetry.RecoverMiddleware(next).ServeHTTP
//	})
func RecoverMiddleware(next http.Handler) http.Handler {
	builder := responsex.New(responsex.Options{
		DefaultErrorCode: errorx.Unknown.Value(),
		DefaultErrorMsg:  errorx.Unknown.Msg(),
		TraceFieldMode:   responsex.TraceFieldModeTraceID,
		TraceFieldExtractor: func(ctx context.Context) string {
			span := trace.SpanFromContext(ctx)
			if span.SpanContext().HasTraceID() {
				return span.SpanContext().TraceID().String()
			}
			return ""
		},
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if span := trace.SpanFromContext(r.Context()); span.IsRecording() {
					span.SetStatus(codes.Error, fmt.Sprintf("panic recovered: %v", v))
				}

				err := errorx.NewCodeMsgDataError(
					errorx.Unknown.Value(),
					"panic recovered: "+fmt.Sprint(v),
					map[string]any{},
				)
				resp := builder.BuildError(r.Context(), err)

				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
