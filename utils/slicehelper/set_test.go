package slicehelper

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSetMap(t *testing.T) {
	type People struct {
		ID   int
		Name string
		Age  int
	}

	var peoples []*People
	for i := 1; i <= 10; i++ {
		peoples = append(peoples, &People{
			ID:   i,
			Name: "zhangSan" + strconv.Itoa(i),
			Age:  rand.Intn(99) + 1,
		})
	}

	namesMap := ToSetMap(peoples, func(i int, elm *People) (bool, string) {
		return true, elm.Name
	})
	assert.Equal(t, 10, len(namesMap))
	assert.True(t, namesMap["zhangSan1"])
	assert.True(t, namesMap["zhangSan5"])
	assert.True(t, namesMap["zhangSan10"])
	assert.False(t, namesMap["zhangSan11"])
}
