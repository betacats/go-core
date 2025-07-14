package restyx

import (
	"fmt"
	"net/http"
	"testing"
)

type RestyResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func TestResty(t *testing.T) {
	client := NewClient()

	var baseUrl string = "https://example.com"

	reqParams := map[string]string{
		"status": "1",
	}
	resp := RestyResp{}
	req := client.R().
		SetHeader("Authorization", "your token").
		SetQueryParams(reqParams).
		SetPathParam("shopID", "1").
		SetResult(resp)

	url := fmt.Sprintf("%s/shop/{shopID}", baseUrl) // -> https://example.com/shop/1?status=1
	ret, err := client.Execute(req, http.MethodGet, url)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(ret.Header().Get("X-Trace-ID"))
	if resp.Code != 200 {
		t.Error(resp.Msg)
		return
	}

	t.Log(resp.Data)
}
