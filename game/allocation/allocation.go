package allocation

import (
	"math/rand"
	"time"

	dip "github.com/zond/godip/common"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	for _, method := range OrderedMethods {
		Methods[method.Name] = method
	}
}

var Methods = map[string]Method{}

var OrderedMethods = []Method{
	Method{
		Name: "Random",
		Allocate: func(nations []dip.Nation, prefs [][]dip.Nation) (result []dip.Nation) {
			result = make([]dip.Nation, len(nations))
			for index, pos := range rand.Perm(len(nations)) {
				result[index] = nations[pos]
			}
			return
		},
	},
	Method{
		Name: "Preferences",
		Allocate: func(nations []dip.Nation, prefs [][]dip.Nation) (result []dip.Nation) {
			return optimizePreferences(prefs)
		},
	},
}

func generatePermutations(prefix, nations []dip.Nation, selected int, result *[][]dip.Nation) {
	if len(prefix) == len(nations) {
		*result = append(*result, prefix)
	} else {
		for index, nation := range nations {
			if selected&(1<<uint(index)) == 0 {
				newPrefix := make([]dip.Nation, len(prefix)+1)
				copy(newPrefix, prefix)
				newPrefix[len(prefix)] = nation
				generatePermutations(newPrefix, nations, selected|(1<<uint(index)), result)
			}
		}
	}
}

func permutations(nations []dip.Nation) (result [][]dip.Nation) {
	generatePermutations(nil, nations, 0, &result)
	return
}

func preferencesScore(perm []dip.Nation, preferences [][]dip.Nation) (result int) {
	for index, chosen := range perm {
		for at, nation := range preferences[index] {
			if nation == chosen {
				result += at * at
				break
			}
		}
	}
	return
}

func optimizePreferences(preferences [][]dip.Nation) (result []dip.Nation) {
	bestScore := 0
	for _, perm := range permutations(preferences[0]) {
		if score := preferencesScore(perm, preferences); result == nil || score < bestScore {
			bestScore = score
			result = perm
		}
	}
	return
}

type Method struct {
	Name     string
	Allocate func(nations []dip.Nation, prefs [][]dip.Nation) (result []dip.Nation)
}
