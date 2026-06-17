package responsex

import "context"

// RequestMeta 表示请求级元数据。
// 业务方可以在 HTTP 中间件中写入，供错误上报器使用。
type RequestMeta struct {
	Method string
	Path   string
	Body   string
	Env    string
}

type requestMetaKey struct{}

// WithRequestMeta 将请求元数据写入上下文。
// 这样后续的错误上报器可以从 ctx 中读取这些信息。
func WithRequestMeta(ctx context.Context, meta RequestMeta) context.Context {
	return context.WithValue(ctx, requestMetaKey{}, meta)
}

// GetRequestMeta 从上下文中读取请求元数据。
// 如果上下文里没有写入，则返回 false。
func GetRequestMeta(ctx context.Context) (RequestMeta, bool) {
	meta, ok := ctx.Value(requestMetaKey{}).(RequestMeta)
	return meta, ok
}
