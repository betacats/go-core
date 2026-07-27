package middlex

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	castx "github.com/betacats/go-core/utils/cast"
	jsoniter "github.com/json-iterator/go"
)

// ArrayParamCompat 将前端 JSON 数组格式的 query 参数转为重复 key 格式。
//
// 前端可能发送 status=[10,20] 这种 JSON 数组作为 query 参数值。
// go-zero 1.6 的 fillSliceFromString 可以 JSON 解析这个值，但 go-zero 1.10
// 的 GetFormValues 返回 []string 后走 fillSlice 无法解析 "[10,20]" 字符串。
//
// 本中间件将 status=[10,20] 转换为 status=10&status=20，兼容以下格式：
//   - JSON 数组：status=[10,20]
//   - 重复 key：status=10&status=20（go-zero 1.10 原生支持，不变）
//   - PHP 括号：status[]=10&status[]=20（go-zero 1.10 原生支持，不变）
//
// 注意：本中间件仅适用于 go-zero 1.10+。如果回退到 go-zero 1.6，需要：
//  1. 删除 server.Use(ArrayParamCompat) 这一行
//  2. 前端 status=[10,20] 格式在 1.6 原生支持，无需中间件
var ArrayParamCompat = func(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery == "" {
			next(w, r)
			return
		}

		q := r.URL.Query()

		// 先快速扫描，无 JSON 数组格式值则跳过，避免不必要的分配
		if !hasArrayValue(q) {
			next(w, r)
			return
		}

		newQuery := make(url.Values)
		for key, values := range q {
			for _, v := range values {
				if v == "" {
					continue
				}
				if v[0] == '[' && v[len(v)-1] == ']' {
					var arr []any
					if err := jsoniter.Unmarshal(castx.StringToBytes(v), &arr); err == nil {
						for _, item := range arr {
							newQuery.Add(key, sprintItem(item))
						}
						continue
					}
				}
				newQuery.Add(key, v)
			}
		}

		r.URL.RawQuery = newQuery.Encode()
		next(w, r)
	}
}

func hasArrayValue(q url.Values) bool {
	for _, values := range q {
		for _, v := range values {
			if v != "" && v[0] == '[' && v[len(v)-1] == ']' {
				return true
			}
		}
	}
	return false
}

func sprintItem(item any) string {
	switch v := item.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	default:
		return fmt.Sprint(item)
	}
}
