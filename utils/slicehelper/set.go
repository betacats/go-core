package slicehelper

// ToSetMap ...
func ToSetMap[T any, Key comparable](elms []T, f func(index int, elm T) (bool, Key)) map[Key]bool {
	res := make(map[Key]bool)
	for i, elm := range elms {
		ok, v := f(i, elm)
		if ok {
			res[v] = true
		}
	}
	return res
}
