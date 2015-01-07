package game

import (
	"reflect"
	"testing"

	dip "github.com/zond/godip/common"
)

func TestOptimizePreferences(t *testing.T) {
	wanted := []dip.Nation{"D", "C", "A", "B"}
	found := optimizePreferences([][]dip.Nation{[]dip.Nation{"D", "A", "B", "C"}, []dip.Nation{"D", "C", "A", "B"}, []dip.Nation{"A", "B", "C", "D"}, []dip.Nation{"A", "B", "C", "D"}})
	if !reflect.DeepEqual(found, wanted) {
		t.Errorf("Wanted %v, but got %v", wanted, found)
	}
}
