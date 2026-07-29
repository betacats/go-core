package middlex

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayParamCompat_Convert(t *testing.T) {
	tests := []struct {
		name     string
		rawQuery string
		want     string
	}{
		{"整数数组", "status=[10,20]", "status=10&status=20"},
		{"单个元素", "status=[10]", "status=10"},
		{"字符串数组", `status=["a","b"]`, "status=a&status=b"},
		{"浮点数数组", "price=[10.5,20.3]", "price=10.5&price=20.3"},
		{"混合类型", `status=[10,"a"]`, "status=10&status=a"},
		{"数组参数与普通参数混合", "status=[10,20]&page=1", "page=1&status=10&status=20"},
		{"多个数组参数", "status=[10,20]&type=[1,2]", "status=10&status=20&type=1&type=2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured string
			handler := ArrayParamCompat(func(w http.ResponseWriter, r *http.Request) {
				captured = r.URL.RawQuery
			})

			req := httptest.NewRequest(http.MethodGet, "/api/list?"+tt.rawQuery, nil)
			handler.ServeHTTP(httptest.NewRecorder(), req)

			// URL 编码后的 key 顺序不保证，按 & split 后比较集合
			assert.ElementsMatch(t, strings.Split(tt.want, "&"), strings.Split(captured, "&"))
		})
	}
}

func TestArrayParamCompat_Passthrough(t *testing.T) {
	tests := []struct {
		name     string
		rawQuery string
	}{
		{"空 query", ""},
		{"普通参数", "page=1&size=10"},
		{"重复 key 格式（原生支持）", "status=10&status=20"},
		{"PHP 括号格式（原生支持）", "status[]=10&status[]=20"},
		{"非法 JSON", "status=[10"},
		{"JSON 对象（非数组）", `filter={"a":1}`},
		{"数字参数字符串（非数组）", "code=[10]x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured string
			handler := ArrayParamCompat(func(w http.ResponseWriter, r *http.Request) {
				captured = r.URL.RawQuery
			})

			url := "/api/list"
			if tt.rawQuery != "" {
				url += "?" + tt.rawQuery
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			handler.ServeHTTP(httptest.NewRecorder(), req)

			assert.Equal(t, tt.rawQuery, captured)
		})
	}
}

func TestArrayParamCompat_EmptyArray(t *testing.T) {
	// 空数组 [] 会被解析为空切片，不添加任何值
	var captured string
	handler := ArrayParamCompat(func(w http.ResponseWriter, r *http.Request) {
		captured = r.URL.RawQuery
	})

	req := httptest.NewRequest(http.MethodGet, "/api/list?status=[]&page=1", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, "page=1", captured)
}

func TestHasArrayValue(t *testing.T) {
	assert.True(t, hasArrayValue(map[string][]string{"status": {"[10,20]"}}))
	assert.True(t, hasArrayValue(map[string][]string{"status": {"[]"}}))
	assert.False(t, hasArrayValue(map[string][]string{"status": {"10"}}))
	assert.False(t, hasArrayValue(map[string][]string{"status": {"[10"}}))
	assert.False(t, hasArrayValue(map[string][]string{}))
}
