package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/betacats/go-core/web/errorx"
	"github.com/betacats/go-core/web/responsex"
)

// Example 演示业务项目在 HTTP handler 中接入 responsex.Builder。
// 这类偏集成式示例放在 examples 子目录，避免挤占 responsex 包根目录。
func Example() {
	type traceKey struct{}

	builder := responsex.New(responsex.Options{
		SuccessCode:    errorx.OK.Value(),
		SuccessMsg:     errorx.OK.Msg(),
		EnableTrace:    true,
		TraceFieldMode: responsex.TraceFieldModeTraceID,
		TraceFieldExtractor: func(ctx context.Context) string {
			v, _ := ctx.Value(traceKey{}).(string)
			return v
		},
		DefaultErrorCode: errorx.Unknown.Value(),
		DefaultErrorMsg:  errorx.Unknown.Msg(),
		EnableReport:     true,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metaErr := responsex.WithRequestMetaFromHTTPRequest(r)
		if metaErr != nil {
			ctx = r.Context()
		}
		ctx = context.WithValue(ctx, traceKey{}, "trace-123")

		if err := createOrder(); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(builder.BuildError(ctx, err))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(builder.BuildOK(ctx, map[string]any{"orderNo": "SO123"}))
	})

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"sku":"dog-food"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Println(strings.TrimSpace(rec.Body.String()))
	// Output:
	// {"result":false,"code":7,"msg":"PERMISSION_DENIED","data":null,"traceId":"trace-123"}
}

func createOrder() error {
	return errorx.NewCodeError(errorx.PermissionDenied)
}
