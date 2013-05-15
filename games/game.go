package games

import (
	"appengine"
	"appengine/datastore"
	"common"
	"fmt"
)

func gameByIdKey(k *datastore.Key) string {
	return fmt.Sprintf("%v{Id:%v}", gameKind, k)
}

const (
	formingGamesKey = "Games{Started:false}"
)

type Games []*Game

type Game struct {
	Id      *datastore.Key `json:"id" datastore:"-"`
	Started bool           `json:"started"`
	Variant string         `json:"variant"`
	EndYear int            `json:"end_year"`
	Private bool           `json:"private"`

	Open bool `json:"open" datastore:"-"`

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
	self.Open = len(self.GetMembers(c)) < len(common.VariantMap[self.Variant].Nations)
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

func (self *Game) Delete(c appengine.Context) (err error) {
	if err = datastore.Delete(c, self.Id); err != nil {
		return
	}
	if !self.Started {
		common.MemDel(c, formingGamesKey)
	}
	common.MemDel(c, gameByIdKey(self.Id))
	return
}

func (self *Game) Save(c appengine.Context, owner string) (result *Game, err error) {
	result = self

	var oldGame *Game
	if self.Id != nil {
		oldGame = GetGameById(c, self.Id)
	}

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
	} else {
		_, err = datastore.Put(c, self.Id, self)
	}

	if oldGame == nil || oldGame.Started != self.Started {
		common.MemDel(c, formingGamesKey)
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

func GetGameById(c appengine.Context, id *datastore.Key) *Game {
	return GetGamesByIds(c, []*datastore.Key{id})[0]
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
