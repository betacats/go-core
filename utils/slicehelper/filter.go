package slicehelper

// Filter 根据传入的匿名函数f返回过滤后的新切片
func Filter[T any](elms []T, f func(index int, elm T) bool) []T {
	var res []T
	for i, v := range elms {
		if f(i, v) {
			res = append(res, v)
		}
	}
	return res
}
