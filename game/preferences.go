package game

import (
	dip "github.com/zond/godip/common"
)

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
