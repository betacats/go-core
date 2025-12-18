package slicehelper

func Reduce[T any, R any](elms []T, f func(cur R, item T) R, init R) R {
	for i := 0; i < len(elms); i++ {
		init = f(init, elms[i])
	}
	return init
}
