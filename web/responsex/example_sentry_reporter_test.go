package responsex_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/betacats/go-core/web/errorx"
	"github.com/betacats/go-core/web/responsex"
)

// ExampleBuilder_withDefaultSentryReporter 演示：开启 EnableReport 后，
// 未显式设置 Reporter 时，Builder 会默认使用内置 SentryReporter。
func ExampleBuilder_withDefaultSentryReporter() {
	builder := responsex.New(responsex.Options{
		SuccessCode:      errorx.OK.Value(),
		SuccessMsg:       errorx.OK.Msg(),
		DefaultErrorCode: errorx.Unknown.Value(),
		DefaultErrorMsg:  errorx.Unknown.Msg(),
		EnableReport:     true,
	})

	ctx := responsex.WithRequestMeta(context.Background(), responsex.RequestMeta{
		Method: "POST",
		Path:   "/orders",
		Body:   `{"sku":"dog-food"}`,
		Env:    "prod",
	})

	resp := builder.BuildError(ctx, errorx.NewCodeError(errorx.PermissionDenied))
	body, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
	// Output:
	// {"result":false,"code":7,"msg":"PERMISSION_DENIED","data":{}}
}

