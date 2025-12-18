package mathx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTernary(t *testing.T) {
	t.Run("", func(t *testing.T) {
		res := Ternary(true, 1, 2)
		assert.Equal(t, 1, res)
	})

	t.Run("", func(t *testing.T) {
		type user struct {
			name string
			age  int
		}
		var u *user
		u = nil
		assert.Panics(t, func() {
			// 会panic
			_ = Ternary(true, u.age, 1)
		})
	})

}

func TestIntersect(t *testing.T) {
	t.Run("", func(t *testing.T) {
		res := Intersect([]int{1, 2, 3}, []int{2, 3, 4})
		assert.ElementsMatch(t, []int{2, 3}, res)
	})

	t.Run("", func(t *testing.T) {
		res := Intersect([]int{1, 2, 3}, []int{4, 5, 6})
		assert.ElementsMatch(t, []int{}, res)
	})

	t.Run("", func(t *testing.T) {
		res := Intersect([]int{}, []int{1, 2, 3})
		assert.ElementsMatch(t, []int{}, res)
	})

	t.Run("", func(t *testing.T) {
		res := Intersect([]int{1, 2, 3}, []int{})
		assert.ElementsMatch(t, []int{}, res)
	})

	t.Run("", func(t *testing.T) {
		res := Intersect([]int{1, 2, 3}, []int{1, 2, 3})
		assert.ElementsMatch(t, []int{1, 2, 3}, res)
	})

	t.Run("", func(t *testing.T) {
		res := Intersect([]int{1, 2, 3}, []int{1, 2, 3, 4})
		assert.ElementsMatch(t, []int{1, 2, 3}, res)
	})
	t.Run("", func(t *testing.T) {
		res := Intersect([]int{}, []int{})
		assert.ElementsMatch(t, []int{}, res)
	})
}

func TestUnion(t *testing.T) {
	t.Run("", func(t *testing.T) {
		res := Union([]int{1, 2, 3}, []int{2, 3, 4})
		assert.ElementsMatch(t, []int{1, 2, 3, 4}, res)
	})

	t.Run("", func(t *testing.T) {
		res := Union([]int{1, 2, 3}, []int{4, 5, 6})
		assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6}, res)
	})

	t.Run("", func(t *testing.T) {
		res := Union([]int{}, []int{1, 2, 3})
		assert.ElementsMatch(t, []int{1, 2, 3}, res)
	})

	t.Run("", func(t *testing.T) {
		res := Union([]int{1, 2, 3}, []int{})
		assert.ElementsMatch(t, []int{1, 2, 3}, res)
	})

	t.Run("", func(t *testing.T) {
		res := Union([]int{1, 2, 3}, []int{1, 2, 3})
		assert.ElementsMatch(t, []int{1, 2, 3}, res)
	})

	t.Run("", func(t *testing.T) {
		res := Union([]int{1, 2, 3}, []int{1, 2, 3, 4})
		assert.ElementsMatch(t, []int{1, 2, 3, 4}, res)
	})
	t.Run("", func(t *testing.T) {
		res := Union([]int{}, []int{})
		assert.ElementsMatch(t, []int{}, res)
	})
}

func TestRemoveEmpty(t *testing.T) {
	t.Run("", func(t *testing.T) {
		res := RemoveEmpty([]int{1, 2, 3, 0})
		assert.ElementsMatch(t, []int{1, 2, 3}, res)
	})
	t.Run("", func(t *testing.T) {
		res := RemoveEmpty([]int{})
		assert.ElementsMatch(t, []int{}, res)
	})
	t.Run("", func(t *testing.T) {
		res := RemoveEmpty([]string{"", "12", "dwqd", "wqdqwd"})
		assert.ElementsMatch(t, []string{"12", "dwqd", "wqdqwd"}, res)
	})
	t.Run("", func(t *testing.T) {
		res := RemoveEmpty([]bool{true, false})
		assert.ElementsMatch(t, []bool{true}, res)
	})
	t.Run("", func(t *testing.T) {
		res := RemoveEmpty([]float64{1.2, 0.0})
		assert.ElementsMatch(t, []float64{1.2}, res)
	})
	t.Run("", func(t *testing.T) {
		res := RemoveEmpty([]struct {
			name string
			age  int
		}{
			{
				name: "1",
				age:  1,
			},
			{
				name: "",
				age:  1,
			},
			{
				name: "12",
				age:  0,
			},
			{
				name: "2",
				age:  2,
			},
			{
				name: "",
				age:  0,
			},
		})
		assert.ElementsMatch(t, []struct {
			name string
			age  int
		}{{
			name: "1",
			age:  1,
		},
			{
				name: "",
				age:  1,
			},
			{
				name: "12",
				age:  0,
			},
			{
				name: "2",
				age:  2,
			}}, res)
	})
}

func TestDifference(t *testing.T) {
	t.Run("", func(t *testing.T) {
		res := Difference([]int{1, 2, 3}, []int{2, 3, 4})
		assert.ElementsMatch(t, []int{1}, res)
	})

	t.Run("", func(t *testing.T) {
		res := Difference([]int{1, 2, 3}, []int{4, 5, 6})
		assert.ElementsMatch(t, []int{1, 2, 3}, res)
	})

	t.Run("", func(t *testing.T) {
		res := Difference([]int{}, []int{1, 2, 3})
		assert.ElementsMatch(t, []int{}, res)
	})
	t.Run("", func(t *testing.T) {
		res := Difference([]int{}, []int{})
		assert.ElementsMatch(t, []int{}, res)
	})
}

func TestSub(t *testing.T) {
	t.Run("", func(t *testing.T) {
		res := Sub(1.2, 0.11)
		assert.Equal(t, 1.09, res)
	})
	t.Run("", func(t *testing.T) {
		res := Sub(1.2, 0.0)
		assert.Equal(t, 1.2, res)
	})
	t.Run("", func(t *testing.T) {
		res := Sub(0.0, 0.1)
		assert.Equal(t, -0.1, res)
	})
}

func TestMul(t *testing.T) {

	t.Run("", func(t *testing.T) {
		res := Mul(1.2, 0.11)
		assert.Equal(t, 0.132, res)
	})

	t.Run("", func(t *testing.T) {
		res := Mul(1.2, 0.0)
		assert.Equal(t, 0.0, res)
	})

	t.Run("", func(t *testing.T) {
		res := Mul(0.0, 0.1)
		assert.Equal(t, 0.0, res)
	})

	t.Run("", func(t *testing.T) {
		res := Mul(0.0, 0.0)
		assert.Equal(t, 0.0, res)
	})

	t.Run("", func(t *testing.T) {
		res := Mul(1.44, 1.44)
		assert.Equal(t, 2.0736, res)
	})

}

func TestDiv(t *testing.T) {
	t.Run("", func(t *testing.T) {
		res := RoundIndex(Div(1.2, 0.11), 0)
		assert.Equal(t, float64(11), res)
	})

	t.Run("", func(t *testing.T) {
		res := RoundIndex(Div(1.2, 0.11), 2)
		assert.Equal(t, 10.91, res)
	})

	t.Run("", func(t *testing.T) {
		res := RoundIndex(Div(1.2, 0.11), 5)
		assert.Equal(t, 10.90909, res)
	})

	t.Run("", func(t *testing.T) {
		res := Div(1.2, 0.0)
		assert.Equal(t, 0.0, res)
	})

	t.Run("", func(t *testing.T) {
		res := Div(0.0, 0.1)
		assert.Equal(t, 0.0, res)
	})

	t.Run("", func(t *testing.T) {
		res := Div(0.0, 0.0)
		assert.Equal(t, 0.0, res)
	})

	t.Run("", func(t *testing.T) {
		res := Div(1.44, 1.44)
		assert.Equal(t, 1.0, res)
	})
}

func TestFloor(t *testing.T) {
	t.Run("", func(t *testing.T) {
		res := Floor(1.2)
		assert.Equal(t, 1.0, res)
	})
	t.Run("", func(t *testing.T) {
		res := Floor(1.0)
		assert.Equal(t, 1.0, res)
	})
	t.Run("", func(t *testing.T) {
		res := Floor(0.0)
		assert.Equal(t, 0.0, res)
	})
	t.Run("", func(t *testing.T) {
		res := Floor(1.5)
		assert.Equal(t, 1.0, res)
	})
	t.Run("", func(t *testing.T) {
		res := Floor(1.99999)
		assert.Equal(t, 1.0, res)
	})
	t.Run("", func(t *testing.T) {
		res := Floor(-1.99999)
		assert.Equal(t, -2.0, res)
	})
}
