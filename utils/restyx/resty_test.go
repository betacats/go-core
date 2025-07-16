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

	resp := RestyResp{}
	req := client.R().
		SetHeader("Authorization", "your token").
		SetQueryParams(map[string]string{
			"status": "1",
		}).
		SetFormData(map[string]string{
			"name": "张三",
		}).
		SetBody(map[string]interface{}{
			"age": 10,
		}).
		SetPathParam("shopID", "1").
		SetResult(resp)

	url := fmt.Sprintf("%s/shop/{shopID}", baseUrl) // -> https://example.com/shop/1?status=1
	ret, err := client.Execute(req, http.MethodPost, url)
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
