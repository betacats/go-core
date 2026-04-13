package slicehelper

// Map ...
func Map[T any, Field any](elms []T, f func(elm T) (bool, Field)) []Field {
	var res []Field
	for _, elm := range elms {
		ok, v := f(elm)
		if ok {
			res = append(res, v)
		}
	}
	return res
}

func ToMap[T any, Key comparable, Value any](elms []T, f func(index int, elm T) (Key, Value)) map[Key]Value {
	res := make(map[Key]Value)
	for index, value := range elms {
		k, v := f(index, value)
		res[k] = v
	}
	return res
}
