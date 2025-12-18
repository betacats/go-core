package mathx

import (
	"github.com/shopspring/decimal"
	"golang.org/x/exp/maps"
)

// Ternary 三目运算符
// 注意：会在调用函数之前对所有参数进行求值，因此无论条件如何，`trueValue` 和 `falseValue` 都会被求值。
// 提及空指针解引用和副作用等需要注意的具体情况。
func Ternary[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	} else {
		return falseValue
	}
}

// Intersect 数组交集
func Intersect[T comparable](a, b []T) []T {
	m := make(map[T]bool)
	for _, v := range a {
		m[v] = true
	}
	var rs = make(map[T]struct{})
	for _, v := range b {
		if m[v] {
			rs[v] = struct{}{}
		}
	}
	return maps.Keys(rs)
}

// Union 数组并集
func Union[T comparable](a, b []T) []T {
	m := make(map[T]bool)
	for _, v := range a {
		m[v] = true
	}
	for _, v := range b {
		m[v] = true
	}
	return maps.Keys(m)
}

// RemoveEmpty 用于可比较类型的切片去除零值
func RemoveEmpty[T comparable](slice []T) []T {
	var result []T
	var zero T
	for _, item := range slice {
		if item != zero {
			result = append(result, item)
		}
	}
	return result
}

// 差集函数，求取 arr1 相对于 arr2 的差集
// arr1-arr2

func Difference[T comparable](arr1, arr2 []T) []T {
	diffMap := make(map[T]bool)
	for _, v := range arr2 {
		diffMap[v] = true
	}
	var result []T
	for _, v := range arr1 {
		if !diffMap[v] {
			result = append(result, v)
		}
	}
	return result
}

// Sub 减法
func Sub(a, b float64) float64 {
	rs, _ := decimal.NewFromFloat(a).Sub(decimal.NewFromFloat(b)).Float64()
	return rs
}

// Mul 乘法
func Mul(a, b float64) float64 {
	rs, _ := decimal.NewFromFloat(a).Mul(decimal.NewFromFloat(b)).Float64()
	return rs
}

// Div 除法
func Div(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	rs, _ := decimal.NewFromFloat(a).Div(decimal.NewFromFloat(b)).Float64()
	return rs
}

// Add 加法
func Add(a, b float64) float64 {
	rs, _ := decimal.NewFromFloat(a).Add(decimal.NewFromFloat(b)).Float64()
	return rs
}

// Adds 加法
func Adds(a float64, b ...float64) float64 {
	sum := decimal.NewFromFloat(a)
	for _, v := range b {
		sum = sum.Add(decimal.NewFromFloat(v))
	}
	rs, _ := sum.Float64()
	return rs
}

// Floor 取整(向下取整)
func Floor(a float64) float64 {
	rs, _ := decimal.NewFromFloat(a).Floor().Float64()
	return rs
}

// Round 取整(四舍五入)
func Round(a float64) float64 {
	rs, _ := decimal.NewFromFloat(a).Round(0).Float64()
	return rs
}

// RoundIndex 截取a小数点后index位(四舍五入)
func RoundIndex(a float64, index int32) float64 {
	rs, _ := decimal.NewFromFloat(a).Round(index).Float64()
	return rs
}

// ExportRound 分转换成元，同时保留小数点后两位
func ExportRound(a float64) float64 {
	rs, _ := decimal.NewFromFloat(a).Div(decimal.NewFromFloat(100)).Round(2).Float64()
	return rs
}
