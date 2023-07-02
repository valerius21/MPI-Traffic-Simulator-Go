package streets

import (
	"fmt"
	"testing"

	"github.com/cornelk/hashmap"
)

const MAX = 1_000

func TestHashMap(t *testing.T) {
	hm := hashmap.New[int, int]()

	for i := 0; i < MAX; i++ {
		hm.Insert(i, i*i)
		fmt.Println(i)
	}
}

func BenchmarkHashMap(b *testing.B) {
	hm := hashmap.New[int, int]()
	for i := 0; i < b.N; i++ {
		hm.Insert(i, i*i)
	}
}
