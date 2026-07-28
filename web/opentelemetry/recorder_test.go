package opentelemetry

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestBodyRecorder_Write 验证 BodyRecorder 同时写入缓冲区和原始 ResponseWriter。
func TestBodyRecorder_Write(t *testing.T) {
	w := httptest.NewRecorder()
	rec := &BodyRecorder{ResponseWriter: w}

	data := []byte(`{"result":true,"code":0,"msg":"success"}`)
	_, err := rec.Write(data)
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if !bytes.Equal(rec.Body(), data) {
		t.Fatalf("body mismatch: got %s, want %s", rec.Body(), data)
	}
	if !bytes.Equal(w.Body.Bytes(), data) {
		t.Fatalf("original response body mismatch: got %s, want %s", w.Body.Bytes(), data)
	}
}

// TestIsBusinessError_ResultFalse 验证 "result":false 被识别为业务错误。
func TestIsBusinessError_ResultFalse(t *testing.T) {
	body := []byte(`{"result":false,"code":7,"msg":"PERMISSION_DENIED"}`)
	if !IsBusinessError(body) {
		t.Fatal("expected business error for result=false")
	}
}

// TestIsBusinessError_CodeNonZero 验证 "code"!=0 被识别为业务错误。
func TestIsBusinessError_CodeNonZero(t *testing.T) {
	body := []byte(`{"result":true,"code":7,"msg":"error"}`)
	if !IsBusinessError(body) {
		t.Fatal("expected business error for code!=0")
	}
}

// TestIsBusinessError_Success 验证成功响应不被识别为业务错误。
func TestIsBusinessError_Success(t *testing.T) {
	body := []byte(`{"result":true,"code":0,"msg":"success"}`)
	if IsBusinessError(body) {
		t.Fatal("expected no business error for success response")
	}
}

// TestIsBusinessError_EmptyBody 验证空 body 不 panic 且不视为错误。
func TestIsBusinessError_EmptyBody(t *testing.T) {
	if IsBusinessError(nil) {
		t.Fatal("expected no business error for nil body")
	}
	if IsBusinessError([]byte{}) {
		t.Fatal("expected no business error for empty body")
	}
}

// TestIsBusinessError_InvalidJSON 验证非法 JSON 不 panic 且不视为错误。
func TestIsBusinessError_InvalidJSON(t *testing.T) {
	if IsBusinessError([]byte(`not json`)) {
		t.Fatal("expected no business error for invalid json")
	}
}

// TestMiddleware_Success 验证成功请求不会标记 span 错误。
func TestMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"result":true,"code":0,"msg":"success"}`))
	})

	wrapped := RequestMiddleware(handler)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	wrapped.ServeHTTP(w, r)

	if !bytes.Equal(w.Body.Bytes(), []byte(`{"result":true,"code":0,"msg":"success"}`)) {
		t.Fatalf("unexpected response: %s", w.Body.String())
	}
}
