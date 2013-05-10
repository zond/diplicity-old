package games

import (
	"appengine"
	"appengine/datastore"
	"common"
	"fmt"
)

const (
	game           = "Game"
	allGamesKey    = "Games{All}"
	gameMemberKind = "GameMember"
	gameKind       = "Game"
	phaseKind      = "Phase"
)

func gameMembersByUserKey(k string) string {
	return fmt.Sprintf("%v{User:%v}", gameMemberKind, k)
}

func gameByIdKey(k *datastore.Key) string {
	return fmt.Sprintf("%v{Id:%v}", k)
}

func latestPhaseByGameIdKey(k *datastore.Key) string {
	return fmt.Sprintf("%v{Game:%v,Latest}", phaseKind, k)
}

type Phases []*Phase

type Phase struct {
	Id      *datastore.Key `json:"id"`
	Season  string         `json:"season"`
	Year    int            `json:"year"`
	Type    int            `json:"type"`
	Ordinal int            `json:"ordinal"`
}

func findLatestPhaseByGameId(c appengine.Context, gameId *datastore.Key) *Phase {
	var phases Phases
	ids, err := datastore.NewQuery(phaseKind).Ancestor(gameId).Order("Ordinal<").Limit(1).GetAll(c, &phases)
	common.AssertOkError(err)
	for index, id := range ids {
		phases[index].Id = id
	}
	if len(phases) == 0 {
		return nil
	}
	return phases[0]
}

func GetLatestPhasesByGameIds(c appengine.Context, gameIds []*datastore.Key) (result Phases) {
	cacheKeys := make([]string, len(gameIds))
	values := make([]interface{}, len(gameIds))
	funcs := make([]func() interface{}, len(gameIds))
	for index, id := range gameIds {
		cacheKeys[index] = latestPhaseByGameIdKey(id)
		idCopy := id
		funcs[index] = func() interface{} {
			return findLatestPhaseByGameId(c, idCopy)
		}
	}
	common.MemoizeMulti(c, cacheKeys, values, funcs)
	result = make(Phases, len(gameIds))
	for index, value := range values {
		result[index] = value.(*Phase)
	}
	return
}

type GameMembers []*GameMember

type GameMember struct {
	Id     *datastore.Key `json:"id" datastore:"-"`
	GameId *datastore.Key `json:"game_id"`
	Nation string         `json:"nation"`

	Game  *Game  `json:"game" datastore:"-"`
	Phase *Phase `json:"phase" datastore:"-"`
}

func findGameMembersByUser(c appengine.Context, email string) (result GameMembers) {
	ids, err := datastore.NewQuery(gameMemberKind).Ancestor(common.UserRoot(c, email)).GetAll(c, &result)
	common.AssertOkError(err)
	for index, _ := range ids {
		result[index].Id = ids[index]
	}
	return
}

func GetGameMembersByUser(c appengine.Context, email string) (result GameMembers) {
	common.Memoize(c, gameMembersByUserKey(email), &result, func() interface{} {
		return findGameMembersByUser(c, email)
	})
	gameIds := make([]*datastore.Key, len(result))
	for index, gameMember := range result {
		gameIds[index] = gameMember.GameId
	}
	for index, game := range GetGamesByIds(c, gameIds) {
		result[index].Game = game
	}
	for index, phase := range GetLatestPhasesByGameIds(c, gameIds) {
		result[index].Phase = phase
	}
	if result == nil {
		result = make(GameMembers, 0)
	}
	return
}

type Games []*Game

type Game struct {
	Id      *datastore.Key `json:"id" datastore:"-"`
	Variant string         `json:"variant"`
}

func findGameById(c appengine.Context, id *datastore.Key) (result Game) {
	err := datastore.Get(c, id, &result)
	common.AssertOkError(err)
	result.Id = id
	return
}

func GetGamesByIds(c appengine.Context, ids []*datastore.Key) (result Games) {
	cacheKeys := make([]string, len(ids))
	values := make([]interface{}, len(ids))
	funcs := make([]func() interface{}, len(ids))
	for index, id := range ids {
		cacheKeys[index] = gameByIdKey(id)
		idCopy := id
		funcs[index] = func() interface{} {
			return findGameById(c, idCopy)
		}
	}
	common.MemoizeMulti(c, cacheKeys, values, funcs)
	result = make(Games, len(ids))
	for index, value := range values {
		result[index] = value.(*Game)
	}
	return
}
