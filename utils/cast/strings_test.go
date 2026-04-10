package castx

import "testing"

func BenchmarkStringToBytes(b *testing.B) {
	s := "这是一段测试文本这是一段测试文本这是一段测试文本这是一段测试文本"
	for i := 0; i < b.N; i++ {
		_ = StringToBytes(s)
	}
}

func BenchmarkStringToBytesStd(b *testing.B) {
	s := "这是一段测试文本这是一段测试文本这是一段测试文本这是一段测试文本"
	for i := 0; i < b.N; i++ {
		_ = []byte(s)
	}
}

//cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
//BenchmarkStringToBytes
//BenchmarkStringToBytes-16       	1000000000	         0.2553 ns/op
//BenchmarkStringToBytesStd
//BenchmarkStringToBytesStd-16    	37896876	        31.39 ns/op
