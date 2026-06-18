package responsex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/betacats/go-core/utils/envx"
)

// ReadRequestBody 读取并恢复 request body。
// 返回 body 字符串和错误信息；读取后会把 body 放回 request，避免影响后续 handler 再次读取。
func ReadRequestBody(r *http.Request) (string, error) {
	if r == nil {
		return "", fmt.Errorf("responsex: request is nil")
	}
	if r.Body == nil {
		return "", nil
	}

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	// 恢复 body，确保后续链路还能正常读取。
	r.Body = io.NopCloser(bytes.NewBuffer(buf))
	return string(buf), nil
}

// WithRequestMetaFromHTTPRequest 从 http.Request 提取元信息并写入 context。
// 这样 Builder.BuildError(ctx, err) 的 Reporter 即使只拿到 ctx，也能获得 method/path/body/env。
func WithRequestMetaFromHTTPRequest(r *http.Request) (context.Context, error) {
	if r == nil {
		return context.Background(), fmt.Errorf("responsex: request is nil")
	}

	body, err := ReadRequestBody(r)
	if err != nil {
		return r.Context(), err
	}

	return WithRequestMeta(r.Context(), RequestMeta{
		Method: r.Method,
		Path:   requestPath(r),
		Body:   body,
		Env:    envx.ENV(),
	}), nil
}

func requestPath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	return r.URL.Path
}

