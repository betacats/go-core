package responsex

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadRequestBodyRestore(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"sku":"dog-food"}`))

	body, err := ReadRequestBody(req)
	if err != nil {
		t.Fatalf("ReadRequestBody err: %v", err)
	}
	if body != `{"sku":"dog-food"}` {
		t.Fatalf("unexpected body: %s", body)
	}

	restored, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read restored body err: %v", err)
	}
	if string(restored) != body {
		t.Fatalf("expected restored body %s, got %s", body, string(restored))
	}
}

func TestWithRequestMetaFromHTTPRequest(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPut, "/v1/coupons/apply", strings.NewReader(`{"coupon":"A100"}`))

	ctx, err := WithRequestMetaFromHTTPRequest(req)
	if err != nil {
		t.Fatalf("WithRequestMetaFromHTTPRequest err: %v", err)
	}

	meta, ok := GetRequestMeta(ctx)
	if !ok {
		t.Fatal("expected request meta in context")
	}
	if meta.Method != http.MethodPut {
		t.Fatalf("expected method PUT, got %s", meta.Method)
	}
	if meta.Path != "/v1/coupons/apply" {
		t.Fatalf("expected path /v1/coupons/apply, got %s", meta.Path)
	}
	if meta.Body != `{"coupon":"A100"}` {
		t.Fatalf("expected body to be preserved, got %s", meta.Body)
	}

	restored, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read restored body err: %v", err)
	}
	if string(restored) != meta.Body {
		t.Fatalf("expected restored body %s, got %s", meta.Body, string(restored))
	}
}

func TestWithRequestMetaFromHTTPRequestNil(t *testing.T) {
	t.Parallel()

	ctx, err := WithRequestMetaFromHTTPRequest(nil)
	if err == nil {
		t.Fatal("expected error when request is nil")
	}
	if ctx == nil {
		t.Fatal("expected non-nil fallback context")
	}
}

