package slicehelper

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type People struct {
	ID   int
	Name string
	Age  int
}

func getPeoples() []*People {
	var peoples []*People
	for i := 1; i <= 10; i++ {
		peoples = append(peoples, &People{
			ID:   i,
			Name: "zhangSan" + strconv.Itoa(i),
			Age:  rand.Intn(99) + 1,
		})
	}
	return peoples
}

func TestMap(t *testing.T) {
	var peoples = getPeoples()

	t.Run("string", func(t *testing.T) {
		names := Map(peoples, func(p *People) (bool, string) {
			return true, p.Name
		})
		assert.Equal(t, []string{"zhangSan1", "zhangSan2", "zhangSan3", "zhangSan4", "zhangSan5", "zhangSan6", "zhangSan7", "zhangSan8", "zhangSan9", "zhangSan10"}, names)
	})

	t.Run("int", func(t *testing.T) {
		ids := Map(peoples, func(p *People) (bool, int) {
			return true, p.ID
		})
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, ids)
	})

	t.Run("struct", func(t *testing.T) {
		newPeoples := Map(peoples, func(p *People) (bool, *People) {
			newp := *p
			newp.Age++
			return true, &newp
		})
		assert.Equal(t, peoples[0].Age+1, newPeoples[0].Age)
	})
}

func TestToMap(t *testing.T) {
	var peoples = getPeoples()

	peoplesNameMap := ToMap(peoples, func(index int, elm *People) (string, *People) {
		return elm.Name, elm
	})

	assert.Equal(t, 10, len(peoplesNameMap))
}
