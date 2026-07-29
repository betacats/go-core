package opentelemetry

import "math/rand"

// sample 以 rate 概率返回 true，用于独立 span 级别的采样决策。
func sample(rate float64) bool {
	if rate <= 0 {
		return false
	}
	if rate >= 1 {
		return true
	}
	return rand.Float64() < rate
}
