package slicehelper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReduce(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		numbs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		sum := Reduce(numbs, func(cur int, item int) int {
			return cur + item
		}, 0)

		assert.Equal(t, 55, sum)
	})

	t.Run("struct", func(t *testing.T) {
		type Student struct {
			Sno   int
			Score float32
		}
		var students []*Student
		for i := 1; i <= 10; i++ {
			students = append(students, &Student{
				Sno:   1000 + i,
				Score: 70 + float32(i),
			})
		}

		sum := Reduce(students, func(cur float32, item *Student) float32 {
			return cur + item.Score
		}, 0)

		assert.Equal(t, float32(755), sum)
	})

	t.Run("string", func(t *testing.T) {
		s := []string{"h", "e", "l", "l", "o", ",", " ", "w", "o", "r", "l", "d"}

		res := Reduce(s, func(cur string, item string) string {
			return cur + item
		}, "")

		assert.Equal(t, "hello, world", res)
	})
}
