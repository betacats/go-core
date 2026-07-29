package middlex

import (
	"encoding/json"
	"net/http"

	"github.com/betacats/go-core/web/responsex"
)

// Recovery 是一个 HTTP 中间件，基于 responsex.Builder 实现 panic 恢复
// 签名风格和 ArrayParamCompat 保持一致，可直接用于 server.Use()
//
// 功能：
//  1. 自动捕获 handler 中的 panic
//  2. 是否上报 sentry 完全由 Builder 的 EnableReport 配置控制
//  3. 复用 Builder 的 Reporter（默认 SentryReporter，支持自定义）
//  4. 返回和普通业务错误完全一致的统一 JSON 响应格式
//
// 使用方式：
//
//	builder := responsex.New(responsex.Options{
//	    EnableReport: true, // 这里控制是否上报 sentry
//	    // ... 其他配置
//	})
//	server.Use(middlex.Recovery(builder))
func Recovery(b *responsex.Builder) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if result := recover(); result != nil {
					// 包装成 PanicError
					panicErr := responsex.NewPanicError(result)

					// 复用 Builder 构建错误响应
					// 这一步会自动：
					// - 根据 EnableReport 配置决定是否上报
					// - 使用配置的 Reporter 上报（默认 SentryReporter）
					// - 生成统一格式的错误响应
					resp := b.BuildError(r.Context(), panicErr)

					// 设置 500 状态码
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					// 写回响应
					_ = json.NewEncoder(w).Encode(resp)
				}
			}()
			next(w, r)
		}
	}
}
