package responsex_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/betacats/go-core/web/errorx"
	"github.com/betacats/go-core/web/responsex"
)

func newBusinessErrorBuilder() *responsex.Builder {
	// 这里没有显式传 Decoder。
	// responsex.New 会默认启用 DecodeResponseCoder()，
	// 而 errorx.CodeError 已经实现了 ResponseCode/ResponseMsg，
	// 所以 BuildError 可以直接把它解析成统一响应。
	return responsex.New(responsex.Options{
		SuccessCode:      errorx.OK.Value(),
		SuccessMsg:       errorx.OK.Msg(),
		DefaultErrorCode: errorx.Unknown.Value(),
		DefaultErrorMsg:  errorx.Unknown.Msg(),
	})
}

func TestBuildErrorWithErrorxCodeError(t *testing.T) {
	t.Parallel()

	builder := newBusinessErrorBuilder()
	resp := builder.BuildError(context.Background(), errorx.NewCodeError(errorx.PermissionDenied))

	if resp.Result != responsex.ResultFailure {
		t.Fatalf("expected failure result, got %v", resp.Result)
	}
	if resp.Code != errorx.PermissionDenied.Value() {
		t.Fatalf("expected code %d, got %d", errorx.PermissionDenied.Value(), resp.Code)
	}
	if resp.Msg != errorx.PermissionDenied.Msg() {
		t.Fatalf("expected msg %q, got %q", errorx.PermissionDenied.Msg(), resp.Msg)
	}
	data, ok := resp.Data.(map[string]any)
	if !ok || len(data) != 0 {
		t.Fatalf("expected default empty object data, got %#v", resp.Data)
	}
}

func TestBuildErrorWithErrorxCodeZeroFallback(t *testing.T) {
	t.Parallel()

	builder := newBusinessErrorBuilder()
	resp := builder.BuildError(context.Background(), errorx.NewCodeMsgError(errorx.OK.Value(), "legacy bad request"))

	if resp.Code != errorx.Unknown.Value() {
		t.Fatalf("expected zero failure code to fallback to unknown=%d, got %d", errorx.Unknown.Value(), resp.Code)
	}
	if resp.Msg != "legacy bad request" {
		t.Fatalf("expected original message to be preserved, got %q", resp.Msg)
	}
}

const (
	couponInvalid  errorx.Code = 990001
	deprecatedCode errorx.Code = 990002
)

// 初始化业务code并注册到 errorx 中，确保它能被 Builder 正确解析。
func InitErrorCode() {

	// 注册 MustRegisterCode
	errorx.MustRegisterCode(couponInvalid, "优惠券无效")
	errorx.MustRegisterCode(deprecatedCode, "已弃用")
}

func TestBuildErrorWithRegisteredBusinessCode(t *testing.T) {
	InitErrorCode()
	builder := newBusinessErrorBuilder()
	resp := builder.BuildError(context.Background(), errorx.NewCodeError(couponInvalid))

	fmt.Println(resp)
	if resp.Code != couponInvalid.Value() {
		t.Fatalf("expected code %d, got %d", couponInvalid.Value(), resp.Code)
	}
	if resp.Msg != couponInvalid.Msg() {
		t.Fatalf("expected msg %q, got %q", couponInvalid.Msg(), resp.Msg)
	}
}

func ExampleBuilder_withErrorxCodeError() {
	builder := newBusinessErrorBuilder()
	resp := builder.BuildError(context.Background(), errorx.NewCodeError(errorx.PermissionDenied))

	body, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
	// Output:
	// {"result":false,"code":7,"msg":"PERMISSION_DENIED","data":{}}
}

func ExampleBuilder_withRegisteredBusinessCode() {
	const couponInvalid errorx.Code = 990003

	if err := errorx.RegisterCode(couponInvalid, "优惠券已失效"); err != nil {
		panic(err)
	}

	builder := newBusinessErrorBuilder()
	resp := builder.BuildError(context.Background(), errorx.NewCodeError(couponInvalid))

	body, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
	// Output:
	// {"result":false,"code":990003,"msg":"优惠券已失效","data":{}}
}
