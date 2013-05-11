package games

import (
	"appengine"
	"appengine/datastore"
	"common"
	"fmt"
)

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
	var phases []Phase
	ids, err := datastore.NewQuery(phaseKind).Ancestor(gameId).Order("Ordinal<").Limit(1).GetAll(c, &phases)
	common.AssertOkError(err)
	for index, id := range ids {
		phases[index].Id = id
	}
	if len(phases) == 0 {
		return nil
	}
	return &phases[0]
}

func GetLatestPhasesByGameIds(c appengine.Context, gameIds []*datastore.Key) (result Phases) {
	cacheKeys := make([]string, len(gameIds))
	values := make([]interface{}, len(gameIds))
	funcs := make([]func() interface{}, len(gameIds))
	for index, id := range gameIds {
		var phase Phase
		values[index] = &phase
		cacheKeys[index] = latestPhaseByGameIdKey(id)
		idCopy := id
		funcs[index] = func() interface{} {
			return findLatestPhaseByGameId(c, idCopy)
		}
	}
	existed := common.MemoizeMulti(c, cacheKeys, values, funcs)
	result = make(Phases, len(gameIds))
	for index, value := range values {
		if existed[index] {
			result[index] = value.(*Phase)
		}
	}
	return
}

type GameMembers []*GameMember

type GameMember struct {
	Id     *datastore.Key `json:"id" datastore:"-"`
	GameId *datastore.Key `json:"game_id"`
	Nation string         `json:"nation,omitempty"`

	Game  *Game  `json:"game,omitempty" datastore:"-"`
	Phase *Phase `json:"phase,omitempty" datastore:"-"`
}

func (self *GameMember) CreateWithGame(c appengine.Context, email string) (result *GameMember, err error) {
	result = self
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		if self.Game, err = self.Game.Save(c); err != nil {
			return err
		}
		self.GameId = self.Game.Id
		_, err = self.Save(c, email)
		return err
	}, &datastore.TransactionOptions{XG: true})
	return
}

func (self *GameMember) Save(c appengine.Context, email string) (result *GameMember, err error) {
	result = self
	if self.GameId == nil {
		err = fmt.Errorf("%+v is missing GameId", self)
		return
	}
	if self.Id == nil {
		self.Id, err = datastore.Put(c, datastore.NewKey(c, gameMemberKind, "", 0, common.UserRoot(c, email)), self)
	} else {
		_, err = datastore.Put(c, self.Id, self)
	}
	common.MemDel(c, gameMembersByUserKey(email))
	return
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
		gameCopy := game
		result[index].Game = gameCopy
	}
	for index, phase := range GetLatestPhasesByGameIds(c, gameIds) {
		phaseCopy := phase
		result[index].Phase = phaseCopy
	}
	if result == nil {
		result = make(GameMembers, 0)
	}
	return
}

type Games []*Game

type Game struct {
	Id      *datastore.Key `json:"id" datastore:"-"`
	Started bool           `json:"started"`
	Variant string         `json:"variant"`
	Private bool           `json:"private"`
}

func findFormingGames(c appengine.Context) (result Games) {
	ids, err := datastore.NewQuery(gameKind).Filter("Started=", false).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		result[index].Id = id
	}
	return
}

func GetFormingGamesForUser(c appengine.Context, email string) (result Games) {
	memberMap := make(map[string]bool)
	for _, member := range GetGameMembersByUser(c, email) {
		memberMap[member.GameId.Encode()] = true
	}

	var preResult Games
	common.Memoize(c, formingGamesKey, &preResult, func() interface{} {
		return findFormingGames(c)
	})

	for _, game := range preResult {
		if !memberMap[game.Id.Encode()] {
			result = append(result, game)
		}
	}

	if result == nil {
		result = make(Games, 0)
	}
	return
}

func (self *Game) Save(c appengine.Context) (result *Game, err error) {
	result = self
	if _, ok := common.VariantMap[self.Variant]; !ok {
		err = fmt.Errorf("Unknown variant: %v", self.Variant)
		return
	}
	if self.Id == nil {
		self.Id, err = datastore.Put(c, datastore.NewKey(c, gameKind, "", 0, nil), self)
		common.MemDel(c, formingGamesKey)
	} else {
		_, err = datastore.Put(c, self.Id, self)
	}
	common.MemDel(c, gameByIdKey(self.Id))
	return
}

func findGameById(c appengine.Context, id *datastore.Key) *Game {
	var result Game
	err := datastore.Get(c, id, &result)
	common.AssertOkError(err)
	result.Id = id
	return &result
}

func GetGamesByIds(c appengine.Context, ids []*datastore.Key) (result Games) {
	cacheKeys := make([]string, len(ids))
	values := make([]interface{}, len(ids))
	funcs := make([]func() interface{}, len(ids))
	for index, id := range ids {
		var game Game
		values[index] = &game
		cacheKeys[index] = gameByIdKey(id)
		idCopy := id
		funcs[index] = func() interface{} {
			return findGameById(c, idCopy)
		}
	}
	existed := common.MemoizeMulti(c, cacheKeys, values, funcs)
	result = make(Games, len(ids))
	for index, value := range values {
		if existed[index] {
			result[index] = value.(*Game)
		}
	}
	return
}
