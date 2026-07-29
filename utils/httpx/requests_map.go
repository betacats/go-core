package httpx

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/jsonx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

// Parse parses the request.
func ParseToMap(r *http.Request) map[string]interface{} {
	result := make(map[string]interface{})

	if ParsePathMap(r) != nil {
		result["path"] = ParsePathMap(r)
	}

	if ParseFormMap(r) != nil {
		result["form"] = ParseFormMap(r)
	}

	if ParseHeadersMap(r) != nil {
		result["header"] = ParseHeadersMap(r)
	}

	if ParseJsonMap(r) != nil {
		result["json"] = ParseJsonMap(r)
	}

	return result
}

func ParseJsonMap(r *http.Request) map[string]interface{} {
	ret := make(map[string]interface{})

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logx.Errorf("Error reading request body: %v", err.Error())
		return nil
	}

	reader := ioutil.NopCloser(bytes.NewBuffer(buf))
	r.Body = reader

	jsonx.UnmarshalFromString(string(buf), &ret)

	return ret
}

// ParseHeaders parses the headers request.
func ParseHeadersMap(r *http.Request) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range r.Header {
		if len(v) == 1 {
			m[k] = v[0]
		} else {
			m[k] = v
		}
	}

	return m
}

// ParseForm parses the form request.
func ParseFormMap(r *http.Request) map[string]interface{} {
	if err := r.ParseForm(); err != nil {
		return nil
	}

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		if err != http.ErrNotMultipart {
			return nil
		}
	}

	params := make(map[string]interface{}, len(r.Form))
	for name := range r.Form {
		formValue := r.Form.Get(name)
		if len(formValue) > 0 {
			params[name] = formValue
		}
	}

	return params
}

// ParseHeader parses the request header and returns a map.
func ParseHeaderMap(headerValue string) map[string]string {
	ret := make(map[string]string)
	fields := strings.Split(headerValue, separator)

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if len(field) == 0 {
			continue
		}

		kv := strings.SplitN(field, "=", tokensInAttribute)
		if len(kv) != tokensInAttribute {
			continue
		}

		ret[kv[0]] = kv[1]
	}

	return ret
}

// ParsePath parses the symbols reside in url path.
// Like http://localhost/bag/:name
func ParsePathMap(r *http.Request) map[string]interface{} {
	vars := pathvar.Vars(r)
	m := make(map[string]interface{}, len(vars))
	for k, v := range vars {
		m[k] = v
	}

	return m
}
