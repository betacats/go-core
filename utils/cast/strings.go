package castx

import (
	"reflect"
	"unsafe"
)

// StringToBytes converts a string to bytes without copy.
func StringToBytes(s string) (b []byte) {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	h.Data = (*reflect.StringHeader)(unsafe.Pointer(&s)).Data
	h.Len = len(s)
	h.Cap = len(s)
	return
}

// BytesToString converts a byte array to string without copy.
func BytesToString(b []byte) (s string) {
	h := (*reflect.StringHeader)(unsafe.Pointer(&s))
	h.Data = (*reflect.SliceHeader)(unsafe.Pointer(&b)).Data
	h.Len = len(b)
	return
}

// 传统方式会进行内存拷贝，性能较低
//s := "hello"
//b := []byte(s)  // 发生内存拷贝
//s2 := string(b) // 再次发生内存拷贝

// 零拷贝方式，性能更高
//s := "hello"
//b := cast.StringToBytes(s)  // 无内存拷贝
//s2 := cast.BytesToString(b) // 无内存拷贝
