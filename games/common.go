package games

import (
	"appengine/datastore"
	"fmt"
)

type Minutes int

const (
	game            = "Game"
	formingGamesKey = "Games{Started:false}"
	gameMemberKind  = "GameMember"
	gameKind        = "Game"
	phaseKind       = "Phase"
)

func gameMembersByUserKey(k string) string {
	return fmt.Sprintf("%v{User:%v}", gameMemberKind, k)
}

func gameMemberByIdKey(k *datastore.Key) string {
	return fmt.Sprintf("%v{Id:%v}", gameMemberKind, k)
}

func gameMembersByGameKey(k *datastore.Key) string {
	return fmt.Sprintf("%v{GameId:%v}", gameMemberKind, k)
}

func gameByIdKey(k *datastore.Key) string {
	return fmt.Sprintf("%v{Id:%v}", gameKind, k)
}

func latestPhaseByGameIdKey(k *datastore.Key) string {
	return fmt.Sprintf("%v{Game:%v,Latest}", phaseKind, k)
}
