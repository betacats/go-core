package slicehelper

// GenContainsFunc 返回一个用于判断传入切片是否存在某个元素的func
// 适用于循环里多次判断
func GenContainsFunc[T comparable](elms []T) func(v T) bool {
	existMap := make(map[T]bool, len(elms))

	for _, s := range elms {
		existMap[s] = true
	}

	return func(v T) bool {
		return existMap[v]
	}
}

// Contains 适用于单次判断元素是否存在
// golang.org/x/exp/slices  内置有
func Contains[T comparable](elms []T, item T) bool {
	for _, elm := range elms {
		if elm == item {
			return true
		}
	}
	return false
}
