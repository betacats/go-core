package batch

import (
	"errors"
	"github.com/betacats/go-core/utils/mathx"
)

var ErrInvalidBatchSize = errors.New("invalid batch num")

// Process 分批处理函数，处理中如果出错立即返回错误
// @list 需要处理的总数据 @batchSize 分批的大小
func Process[V any](list []V, batchSize int, f func(batchList []V) error) error {
	if batchSize <= 0 {
		return ErrInvalidBatchSize
	}
	var (
		l        = len(list)
		batchNum = mathx.Ternary(l%batchSize > 0, l/batchSize+1, l/batchSize)
	)
	for i := 0; i < batchNum; i++ {
		start, end := batchSize*i, min(batchSize*(i+1), l)
		err := f(list[start:end])
		if err != nil {
			return err
		}
	}
	return nil
}
