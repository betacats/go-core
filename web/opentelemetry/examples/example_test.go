package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/betacats/go-core/web/opentelemetry"
)

// Example 演示 opentelemetry 在 REST 服务中的接入方式。
//
// 实际项目入口（cmd/cobra.go）：
//
//	cfg := opentelemetry.NewConfig(
//	    "api-xd",
//	    "http://arms.aliyuncs.com:443",
//	    "/v1/traces",
//	)
//	cfg.Headers = map[string]string{"ARMS-USERNAME": "u", "ARMS-PASSWORD": "p"}
//	shutdown, err := opentelemetry.InitTracing(cfg)
//	proc.AddShutdownListener(func() { _ = shutdown(context.Background()) })
func Example() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if r.URL.Path == "/orders" && r.Method == http.MethodPost {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result": false,
				"code":   7,
				"msg":    "PERMISSION_DENIED",
				"data":   map[string]any{},
			})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": true,
			"code":   0,
			"msg":    "success",
			"data":   map[string]any{"orderNo": "SO123"},
		})
	})

	wrapped := opentelemetry.RequestMiddleware(handler)

	reqOK := httptest.NewRequest(http.MethodGet, "/health", nil)
	recOK := httptest.NewRecorder()
	wrapped.ServeHTTP(recOK, reqOK)
	fmt.Println("health:", string(bytes.TrimSpace(recOK.Body.Bytes())))

	reqErr := httptest.NewRequest(http.MethodPost, "/orders", nil)
	recErr := httptest.NewRecorder()
	wrapped.ServeHTTP(recErr, reqErr)
	fmt.Println("order:", string(bytes.TrimSpace(recErr.Body.Bytes())))

	// Output:
	// health: {"code":0,"data":{"orderNo":"SO123"},"msg":"success","result":true}
	// order: {"code":7,"data":{},"msg":"PERMISSION_DENIED","result":false}
}

// ExampleMiddleware 演示 Middleware 对响应体的拦截。
func ExampleMiddleware() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": false,
			"code":   7,
			"msg":    "PERMISSION_DENIED",
		})
	}

	wrapped := opentelemetry.RequestMiddleware(http.HandlerFunc(handler))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	fmt.Println("response:", string(bytes.TrimSpace(rec.Body.Bytes())))

	// Output:
	// response: {"code":7,"msg":"PERMISSION_DENIED","result":false}
}

// ExampleBodyRecorder 演示 BodyRecorder 直接使用。
func ExampleBodyRecorder() {
	rec := &opentelemetry.BodyRecorder{ResponseWriter: httptest.NewRecorder()}
	_, _ = rec.Write([]byte(`{"result":false,"code":7,"msg":"error"}`))

	fmt.Println("is business error:", opentelemetry.IsBusinessError(rec.Body()))
	fmt.Println("body:", string(rec.Body()))

	// Output:
	// is business error: true
	// body: {"result":false,"code":7,"msg":"error"}
}
