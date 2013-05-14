package games

import (
	"appengine"
	"appengine/datastore"
	"common"
	"fmt"
	dip "github.com/zond/godip/common"
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

func gameByIdKey(k *datastore.Key) string {
	return fmt.Sprintf("%v{Id:%v}", gameKind, k)
}

func latestPhaseByGameIdKey(k *datastore.Key) string {
	return fmt.Sprintf("%v{Game:%v,Latest}", phaseKind, k)
}

type Phases []*Phase

type Phase struct {
	Id      *datastore.Key `json:"id"`
	Season  dip.Season     `json:"season"`
	Year    int            `json:"year"`
	Type    dip.PhaseType  `json:"type"`
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
	Nation dip.Nation     `json:"nation,omitempty"`

	Owner bool   `json:"owner" datastore:"-"`
	Game  *Game  `json:"game,omitempty" datastore:"-"`
	Phase *Phase `json:"phase,omitempty" datastore:"-"`
}

func (self *GameMember) CopyFrom(o *GameMember) *GameMember {
	self.Game.CopyFrom(o.Game)
	return self
}

func (self *GameMember) SaveWithGame(c appengine.Context, email string) (result *GameMember, err error) {
	result = self
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		if self.Game, err = self.Game.Save(c, email); err != nil {
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
		if game.Id.Parent().StringID() == email {
			result[index].Owner = true
		}
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

func GetFormingGamesForUser(c appengine.Context, email string) (result GameMembers) {
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
			owner := game.Id.Parent().StringID() == email
			result = append(result, &GameMember{
				Game:  game.process(c),
				Owner: owner,
			})
		}
	}

	if result == nil {
		result = make(GameMembers, 0)
	}
	return
}

func findGameMemberById(c appengine.Context, key *datastore.Key) *GameMember {
	var result GameMember
	err := datastore.Get(c, key, &result)
	common.AssertOkError(err)
	result.Id = key
	return &result
}

func GetGameMemberById(c appengine.Context, key *datastore.Key) *GameMember {
	var result GameMember
	if common.Memoize(c, gameMemberByIdKey(key), &result, func() interface{} {
		return findGameMemberById(c, key)
	}) {
		result.Game = GetGamesByIds(c, []*datastore.Key{result.GameId})[0]
		result.Phase = GetLatestPhasesByGameIds(c, []*datastore.Key{result.GameId})[0]
		return &result
	}
	return nil
}

type Games []*Game

type Game struct {
	Id      *datastore.Key `json:"id" datastore:"-"`
	Started bool           `json:"started"`
	Variant string         `json:"variant"`
	EndYear int            `json:"end_year"`
	Private bool           `json:"private"`

	SerializedDeadlines []byte             `json:"-"`
	Deadlines           map[string]Minutes `json:"deadlines" datastore:"-"`

	SerializedChatFlags []byte                     `json:"-"`
	ChatFlags           map[string]common.ChatFlag `json:"chat_flags" datastore:"-"`
}

func (self *Game) CopyFrom(o *Game) *Game {
	self.Variant = o.Variant
	self.EndYear = o.EndYear
	self.Private = o.Private
	self.Deadlines = o.Deadlines
	self.ChatFlags = o.ChatFlags
	return self
}

func (self *Game) process(c appengine.Context) *Game {
	common.MustUnmarshalJSON(self.SerializedDeadlines, &self.Deadlines)
	common.MustUnmarshalJSON(self.SerializedChatFlags, &self.ChatFlags)
	return self
}

func findFormingGames(c appengine.Context) (result Games) {
	ids, err := datastore.NewQuery(gameKind).Filter("Started=", false).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		result[index].Id = id
	}
	return
}

func (self *Game) Save(c appengine.Context, owner string) (result *Game, err error) {
	result = self

	if self.Deadlines == nil {
		self.Deadlines = make(map[string]Minutes)
	}
	self.SerializedDeadlines = common.MustMarshalJSON(self.Deadlines)

	if self.ChatFlags == nil {
		self.ChatFlags = make(map[string]common.ChatFlag)
	}
	self.SerializedChatFlags = common.MustMarshalJSON(self.ChatFlags)

	if _, ok := common.VariantMap[self.Variant]; !ok {
		err = fmt.Errorf("Unknown variant: %v", self.Variant)
		return
	}
	if self.Id == nil {
		self.Id, err = datastore.Put(c, datastore.NewKey(c, gameKind, "", 0, common.UserRoot(c, owner)), self)
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
			result[index] = value.(*Game).process(c)
		}
	}
	return
}
