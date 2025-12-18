package slicehelper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilter(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		arr := []int{7, -2, 3, -8, 1, 9, 0, -4, 5, 2, -1}

		res := Filter(arr, func(index, elm int) bool {
			return elm >= 0
		})

		assert.Equal(t, []int{7, 3, 1, 9, 0, 5, 2}, res)
	})

	t.Run("struct", func(t *testing.T) {
		type People struct {
			Name string
			Age  int
		}

		peoples := []*People{
			{
				Name: "zhangSan1",
				Age:  11,
			},
			{
				Name: "zhangSan2",
				Age:  12,
			},
			{
				Name: "zhangSan3",
				Age:  13,
			},
			{
				Name: "zhangSan4",
				Age:  14,
			},
			{
				Name: "zhangSan5",
				Age:  15,
			},
		}

		res := Filter(peoples, func(index int, elm *People) bool {
			return elm.Age >= 13
		})

		assert.Equal(t, []*People{
			{
				Name: "zhangSan3",
				Age:  13,
			},
			{
				Name: "zhangSan4",
				Age:  14,
			},
			{
				Name: "zhangSan5",
				Age:  15,
			},
		}, res)
	})

	t.Run("deduplication", func(t *testing.T) {
		arr := []int{7, 7, 3, 3, 1, 9, 0, -4, 5, 2, -1}

		// 使用map来过滤掉重复项
		seen := make(map[int]bool)
		res := Filter(arr, func(index int, elm int) bool {
			if seen[elm] {
				return false
			}
			seen[elm] = true
			return true
		})

		assert.Equal(t, []int{7, 3, 1, 9, 0, -4, 5, 2, -1}, res)
	})
}
