package batch

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcess(t *testing.T) {
	list := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	t.Run("", func(t *testing.T) {
		err := Process(list, 3, func(batchList []int) error {
			fmt.Println(batchList)
			return nil
		})
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		expected := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}}
		var res [][]int
		err := Process(list, 3, func(batchList []int) error {
			res = append(res, batchList)
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("", func(t *testing.T) {
		expected := [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
		var res [][]int
		err := Process(list, 10, func(batchList []int) error {
			res = append(res, batchList)
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("", func(t *testing.T) {
		expected := [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
		var res [][]int
		err := Process(list, 20, func(batchList []int) error {
			res = append(res, batchList)
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("error", func(t *testing.T) {
		err := Process(list, 0, func(batchList []int) error {
			return nil
		})
		assert.NotNil(t, err)
	})
}
