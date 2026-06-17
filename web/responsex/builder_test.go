package responsex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type traceKey struct{}

// mockCodeError 是测试用的业务错误实现。
// 它实现了 ResponseCoder、ResponseResultProvider、ResponseDataProvider。
type mockCodeError struct {
	code   int
	msg    string
	result ResultCode
	data   any
}

// Error 返回 error 标准文案。
func (e *mockCodeError) Error() string {
	return e.msg
}

// ResponseCode 返回响应错误码。
func (e *mockCodeError) ResponseCode() int {
	return e.code
}

// ResponseMsg 返回响应错误文案。
func (e *mockCodeError) ResponseMsg() string {
	return e.msg
}

// ResponseResult 返回业务结果。
func (e *mockCodeError) ResponseResult() ResultCode {
	return e.result
}

// ResponseData 返回错误响应中的 data。
func (e *mockCodeError) ResponseData() any {
	return e.data
}

// TestBuildOK 用于验证成功响应构建逻辑。
func TestBuildOK(t *testing.T) {
	t.Parallel()

	builder := New(Options{
		SuccessCode:    0,
		SuccessMsg:     "success",
		TraceFieldMode: TraceFieldModeTraceID,
		TraceFieldExtractor: func(ctx context.Context) string {
			v, _ := ctx.Value(traceKey{}).(string)
			return v
		},
	})

	ctx := context.WithValue(context.Background(), traceKey{}, "trace-123")
	resp := builder.BuildOK(ctx, map[string]any{"id": 1})

	if resp.Code != 0 {
		t.Fatalf("expected code 200, got %d", resp.Code)
	}
	if resp.Msg != "success" {
		t.Fatalf("expected msg 成功, got %s", resp.Msg)
	}
	if resp.TraceID != "trace-123" {
		t.Fatalf("expected traceId trace-123, got %s", resp.TraceID)
	}
	if resp.Span != "" {
		t.Fatalf("expected empty span, got %s", resp.Span)
	}
}

// TestBuildOKDefaults 用于验证成功响应默认值为 code=0、msg=success。
func TestBuildOKDefaults(t *testing.T) {
	t.Parallel()

	builder := New(Options{})
	resp := builder.BuildOK(context.Background(), nil)

	if resp.Code != 0 {
		t.Fatalf("expected default code 0, got %d", resp.Code)
	}
	if resp.Msg != "success" {
		t.Fatalf("expected default msg success, got %s", resp.Msg)
	}
	data, ok := resp.Data.(map[string]any)
	if !ok || len(data) != 0 {
		t.Fatalf("expected default normalized empty object data, got %#v", resp.Data)
	}
}

// TestResponseMarshalJSONWithoutTraceField 用于验证关闭链路字段时仅输出共识的 4 个字段。
func TestResponseMarshalJSONWithoutTraceField(t *testing.T) {
	t.Parallel()

	builder := New(Options{
		SuccessCode: 0,
		SuccessMsg:  "success",
	})

	body, err := json.Marshal(builder.BuildOK(context.Background(), map[string]any{"id": 1}))
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(payload) != 4 {
		t.Fatalf("expected 4 fields only, got %#v", payload)
	}
	if _, ok := payload["traceId"]; ok {
		t.Fatalf("expected traceId to be absent, got %#v", payload)
	}
	if _, ok := payload["span"]; ok {
		t.Fatalf("expected span to be absent, got %#v", payload)
	}
}

// TestResponseMarshalJSONWithSpanField 用于验证开启 span 模式时的 JSON 输出。
func TestResponseMarshalJSONWithSpanField(t *testing.T) {
	t.Parallel()

	builder := New(Options{
		SuccessCode:    0,
		SuccessMsg:     "success",
		TraceFieldMode: TraceFieldModeSpan,
		TraceFieldExtractor: func(ctx context.Context) string {
			v, _ := ctx.Value(traceKey{}).(string)
			return v
		},
	})

	ctx := context.WithValue(context.Background(), traceKey{}, "span-123")
	body, err := json.Marshal(builder.BuildOK(ctx, map[string]any{"id": 1}))
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	fmt.Printf("payload: %#v\n", payload)

	if payload["span"] != "span-123" {
		t.Fatalf("expected span=span-123, got %#v", payload)
	}
	if _, ok := payload["traceId"]; ok {
		t.Fatalf("expected traceId to be absent when field name is span, got %#v", payload)
	}
	if got := string(body); got != `{"result":true,"code":0,"msg":"success","data":{"id":1},"span":"span-123"}` {
		t.Fatalf("expected ordered json output, got %s", got)
	}
}

// TestResponseMarshalJSONWithoutTraceMode 用于验证未开启链路字段模式时不会输出 traceId/span。
func TestResponseMarshalJSONWithoutTraceMode(t *testing.T) {
	t.Parallel()

	builder := New(Options{
		SuccessCode: 0,
		SuccessMsg:  "success",
		TraceFieldExtractor: func(ctx context.Context) string {
			return "trace-123"
		},
	})

	body, err := json.Marshal(builder.BuildOK(context.Background(), map[string]any{"id": 1}))
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(payload) != 4 {
		t.Fatalf("expected no trace field output without trace mode, got %#v", payload)
	}
}

// TestBuildOKWithStdHTTP 用于验证 Response 直接交给标准 net/http JSON 输出时的最终结构。
func TestBuildOKWithStdHTTP(t *testing.T) {
	t.Parallel()

	builder := New(Options{
		SuccessCode:    0,
		SuccessMsg:     "success",
		TraceFieldMode: TraceFieldModeSpan,
		TraceFieldExtractor: func(ctx context.Context) string {
			v, _ := ctx.Value(traceKey{}).(string)
			return v
		},
	})

	req := httptest.NewRequest("GET", "/demo", nil)
	ctx := context.WithValue(req.Context(), traceKey{}, "span-123")
	req = req.WithContext(ctx)

	resp := builder.BuildOK(req.Context(), map[string]any{"id": 1})

	stdRec := httptest.NewRecorder()
	stdRec.Header().Set("Content-Type", "application/json; charset=utf-8")
	stdRec.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(stdRec).Encode(resp); err != nil {
		t.Fatalf("encode stdlib response: %v", err)
	}

	stdBody := stdRec.Body.String()
	t.Logf("net/http body: %s", stdBody)
	if trimmed := strings.TrimSpace(stdBody); trimmed != `{"result":true,"code":0,"msg":"success","data":{"id":1},"span":"span-123"}` {
		t.Fatalf("expected ordered http json body, got %s", trimmed)
	}

	if stdRec.Code != http.StatusOK {
		t.Fatalf("expected stdlib status 200, got %d", stdRec.Code)
	}

	var stdPayload map[string]any
	if err := json.Unmarshal(stdRec.Body.Bytes(), &stdPayload); err != nil {
		t.Fatalf("unmarshal stdlib response body: %v", err)
	}
	assertHTTPPayload(t, stdPayload)
}

func assertHTTPPayload(t *testing.T, payload map[string]any) {
	t.Helper()

	if payload["result"] != true {
		t.Fatalf("expected result=true, got %#v", payload)
	}
	if payload["msg"] != "success" {
		t.Fatalf("expected msg=success, got %#v", payload)
	}
	if payload["code"] != float64(0) {
		t.Fatalf("expected code=0, got %#v", payload)
	}
	if payload["span"] != "span-123" {
		t.Fatalf("expected span=span-123, got %#v", payload)
	}
	if _, ok := payload["traceId"]; ok {
		t.Fatalf("expected traceId to be absent, got %#v", payload)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok || data["id"] != float64(1) {
		t.Fatalf("expected data.id=1, got %#v", payload)
	}
}

// TestBuildErrorByResponseCoder 用于验证自定义业务错误解析逻辑。
func TestBuildErrorByResponseCoder(t *testing.T) {
	t.Parallel()

	builder := New(Options{
		DefaultErrorCode: 2,
		DefaultErrorMsg:  "未知错误",
	})

	err := &mockCodeError{
		code:   401,
		msg:    "未登录",
		result: ResultFailure,
		data:   map[string]any{"reason": "token expired"},
	}

	resp := builder.BuildError(context.Background(), err)

	if resp.Code != 401 {
		t.Fatalf("expected code 401, got %d", resp.Code)
	}
	if resp.Msg != "未登录" {
		t.Fatalf("expected msg 未登录, got %s", resp.Msg)
	}
	data, ok := resp.Data.(map[string]any)
	if !ok || data["reason"] != "token expired" {
		t.Fatalf("expected response data to carry decoder data, got %#v", resp.Data)
	}
}

// TestBuildErrorFallback 用于验证默认错误兜底逻辑。
func TestBuildErrorFallback(t *testing.T) {
	t.Parallel()

	builder := New(Options{
		DefaultErrorCode: 500,
		DefaultErrorMsg:  "系统异常",
	})

	resp := builder.BuildError(context.Background(), errors.New("db down"))

	if resp.Code != 500 {
		t.Fatalf("expected code 500, got %d", resp.Code)
	}
	if resp.Msg != "db down" {
		t.Fatalf("expected msg db down, got %s", resp.Msg)
	}
}

// TestBuildErrorReport 用于验证错误上报逻辑是否被触发。
func TestBuildErrorReport(t *testing.T) {
	t.Parallel()

	reported := false
	builder := New(Options{
		DefaultErrorCode: 500,
		DefaultErrorMsg:  "系统异常",
		EnableReport:     true,
		Reporter: ReporterFunc(func(ctx context.Context, payload ReporterPayload) {
			reported = true
			if payload.Response.Code != 500 {
				t.Fatalf("expected reported response code 500, got %d", payload.Response.Code)
			}
		}),
	})

	builder.BuildError(context.Background(), errors.New("boom"))

	if !reported {
		t.Fatal("expected reporter to be called")
	}
}

func TestNewDefaultReporterWhenEnableReport(t *testing.T) {
	t.Parallel()

	builder := New(Options{EnableReport: true})
	if builder.opts.Reporter == nil {
		t.Fatal("expected default reporter to be initialized")
	}
	if _, ok := builder.opts.Reporter.(*SentryReporter); !ok {
		t.Fatalf("expected default reporter to be *SentryReporter, got %T", builder.opts.Reporter)
	}
}

func TestBuildErrorShouldReportFalse(t *testing.T) {
	t.Parallel()

	reported := false
	builder := New(Options{
		EnableReport: true,
		Reporter: ReporterFunc(func(ctx context.Context, payload ReporterPayload) {
			reported = true
		}),
		ShouldReport: func(ctx context.Context, parsed ParsedError) bool {
			return false
		},
	})

	builder.BuildError(context.Background(), errors.New("boom"))
	if reported {
		t.Fatal("expected reporter not to be called when ShouldReport returns false")
	}
}

