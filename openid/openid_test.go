package openid

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestOldNonces(t *testing.T) {
	n := newOldNonces()
	n.max = 5
	for i := 0; i < n.max*2; i++ {
		k := fmt.Sprint(rand.Int())
		if !n.add(k) {
			t.Fatalf("%v should be able to accept %v", n, k)
		}
		if n.size() > n.max {
			t.Fatalf("%v should only be %v big", n, n.max)
		}
		if n.add(k) {
			t.Fatalf("%v should not accept %v", n, k)
		}
	}
}
