package games

import (
	"appengine"
	"appengine/datastore"
	"common"
	"fmt"
	"github.com/zond/godip/classical"
	dip "github.com/zond/godip/common"
	"math/rand"
)

func gameByIdKey(k *datastore.Key) string {
	return fmt.Sprintf("%v{Id:%v}", gameKind, k)
}

const (
	openGamesKey = "Games{Closed:false}"
)

type Games []*Game

type Game struct {
	Id      *datastore.Key `json:"id" datastore:"-"`
	Started bool           `json:"started"`
	Closed  bool           `json:"closed"`
	Variant string         `json:"variant"`
	EndYear int            `json:"end_year"`
	Private bool           `json:"private"`
	Owner   string         `json:"owner"`

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

func findOpenGames(c appengine.Context) (result Games) {
	ids, err := datastore.NewQuery(gameKind).Filter("Closed=", false).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		result[index].Id = id
	}
	return
}

func GetOpenGames(c appengine.Context) (result Games) {
	common.Memoize(c, openGamesKey, &result, func() interface{} {
		return findOpenGames(c)
	})
	if result == nil {
		result = make(Games, 0)
	}
	return
}

func (self *Game) Delete(c appengine.Context) (err error) {
	if err = datastore.Delete(c, self.Id); err != nil {
		return
	}
	if !self.Closed {
		common.MemDel(c, openGamesKey)
	}
	common.MemDel(c, gameByIdKey(self.Id))
	return
}

func (self *Game) start(c appengine.Context, members map[string]*GameMember) (err error) {
	self.Started = true
	self.Closed = true
	allNations := common.VariantMap[self.Variant].Nations
	availableNations := make([]dip.Nation, len(allNations))
	copy(availableNations, allNations)
	for _, member := range members {
		chosenIndex := rand.Int() % len(availableNations)
		member.Nation = availableNations[chosenIndex]
		if chosenIndex == 0 {
			availableNations = availableNations[1:]
		} else if chosenIndex == len(availableNations)-1 {
			availableNations = availableNations[:len(availableNations)-1]
		} else {
			availableNations = append(availableNations[:chosenIndex], availableNations[chosenIndex+1:]...)
		}
		if _, err = member.save(c, member.Email, self.Id); err != nil {
			return
		}
	}
	s := classical.Start()
	p := s.Phase()
	phase := &Phase{
		Season:  p.Season(),
		Year:    p.Year(),
		Type:    p.Type(),
		Ordinal: 1,
	}
	phase.Save(c, self.Id)
	_, err = self.save(c, self.Owner)
	return
}

func (self *Game) save(c appengine.Context, owner string) (result *Game, err error) {
	result = self

	var oldGame *Game
	if self.Id != nil {
		oldGame = GetGameById(c, self.Id, false)
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
		self.Id, err = datastore.Put(c, datastore.NewKey(c, gameKind, "", 0, nil), self)
	} else {
		_, err = datastore.Put(c, self.Id, self)
	}

	if oldGame == nil || oldGame.Closed != self.Closed || !self.Closed {
		common.MemDel(c, openGamesKey)
	}
	common.MemDel(c, gameByIdKey(self.Id))
	return
}

func findGameById(c appengine.Context, id *datastore.Key) *Game {
	var result Game
	err := datastore.Get(c, id, &result)
	if err == datastore.ErrNoSuchEntity {
		return nil
	}
	common.AssertOkError(err)
	result.Id = id
	return &result
}

func GetGameById(c appengine.Context, id *datastore.Key, useCache bool) *Game {
	if useCache {
		return GetGamesByIds(c, []*datastore.Key{id})[0]
	}
	return findGameById(c, id).process(c)
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
