package slicehelper

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"math/rand"
	"testing"
)

func TestGenContainsFunc(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	contains := GenContainsFunc(items)

	assert.Equal(t, true, contains(1))
	assert.Equal(t, true, contains(5))
	assert.Equal(t, true, contains(10))
	assert.Equal(t, false, contains(20))
	assert.Equal(t, false, contains(21))
}

func TestContains(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	assert.Equal(t, true, Contains(items, 1))
	assert.Equal(t, true, Contains(items, 5))
	assert.Equal(t, true, Contains(items, 10))
	assert.Equal(t, false, Contains(items, 20))
	assert.Equal(t, false, Contains(items, 21))
}

var slice []int

func init() {
	for i := 0; i < 10000; i++ {
		slice = append(slice, rand.Intn(1000))
	}
}

func BenchmarkSlicesContains(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = slices.Contains(slice, i)
	}
}

func BenchmarkExistMap(b *testing.B) {
	exists := make(map[int]bool, len(slice))
	for _, v := range slice {
		exists[v] = true
	}
	for i := 0; i < b.N; i++ {
		_ = exists[i]
	}
}

func BenchmarkGenContainsFunc(b *testing.B) {
	contains := GenContainsFunc(slice)

	for i := 0; i < b.N; i++ {
		_ = contains(i)
	}
}

//cpu: Apple M4 Pro

//BenchmarkContains
//BenchmarkContains-14    	  528915	      2339 ns/op

//BenchmarkExistMap
//BenchmarkExistMap-14    	366459024	         3.231 ns/op

//BenchmarkGenContainsFunc
//BenchmarkGenContainsFunc-14    	367099287	         3.228 ns/op
