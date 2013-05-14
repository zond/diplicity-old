package games

import (
	"appengine"
	"appengine/datastore"
	"common"
	"fmt"
	dip "github.com/zond/godip/common"
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

type GameMembers []*GameMember

type GameMember struct {
	Id     *datastore.Key `json:"id" datastore:"-"`
	GameId *datastore.Key `json:"game_id"`
	Email  string         `json:"email"`
	Nation dip.Nation     `json:"nation,omitempty"`

	Owner bool   `json:"owner" datastore:"-"`
	Game  *Game  `json:"game,omitempty" datastore:"-"`
	Phase *Phase `json:"phase,omitempty" datastore:"-"`
}

func (self *GameMember) CopyFrom(o *GameMember) *GameMember {
	self.Game.CopyFrom(o.Game)
	return self
}

func (self *GameMember) IdByGame(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, gameMemberKind, self.Email, 0, self.GameId)
}

func (self *GameMember) IdByEmail(c appengine.Context) *datastore.Key {
	rval := datastore.NewKey(c, gameMemberKind, self.GameId.Encode(), 0, common.UserRoot(c, self.Email))
	c.Infof("self: %+v, id: %v", self, rval)
	return rval
}

func (self *GameMember) ValidatedDelete(c appengine.Context, email string) {
	if err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		// validate that user is member
		if self.Email != email {
			return fmt.Errorf("%+v is not %v", self, email)
		}
		var game *Game
		if game = GetGamesByIds(c, []*datastore.Key{self.GameId})[0]; game.Started {
			return fmt.Errorf("%+v has already started", game)
		}
		if err := datastore.Delete(c, self.IdByGame(c)); err != nil {
			return err
		}
		if err := datastore.Delete(c, self.IdByEmail(c)); err != nil {
			return err
		}

		if len(game.GetMembers(c)) == 0 {
			if err := game.Delete(c); err != nil {
				return
			}
		}

		common.MemDel(c, gameMembersByUserKey(email), gameMembersByGameKey(self.GameId), gameMemberByIdKey(self.IdByGame(c)))

		return nil
	}, &datastore.TransactionOptions{XG: true}); err != nil {
		panic(err)
	}
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
	if _, err = datastore.Put(c, self.IdByGame(c), self); err != nil {
		return
	}
	if _, err = datastore.Put(c, self.IdByEmail(c), self); err != nil {
		return
	}

	common.MemDel(c, gameMembersByUserKey(email), gameMembersByGameKey(self.GameId), gameMemberByIdKey(self.IdByGame(c)))
	return
}

func findGameMembersByGameId(c appengine.Context, gameId *datastore.Key) (result GameMembers) {
	_, err := datastore.NewQuery(gameMemberKind).Ancestor(gameId).GetAll(c, &result)
	common.AssertOkError(err)
	return
}

func (self *Game) GetMembers(c appengine.Context) (result GameMembers) {
	common.Memoize(c, gameMembersByGameKey(self.Id), &result, func() interface{} {
		return findGameMembersByGameId(c, self.Id)
	})
	if result == nil {
		result = make(GameMembers, 0)
	}
	return
}

func findGameMembersByUser(c appengine.Context, email string) (result GameMembers) {
	_, err := datastore.NewQuery(gameMemberKind).Ancestor(common.UserRoot(c, email)).GetAll(c, &result)
	common.AssertOkError(err)
	return
}

func GetGameMembersByUser(c appengine.Context, email string) (result GameMembers) {
	common.Memoize(c, gameMembersByUserKey(email), &result, func() interface{} {
		return findGameMembersByUser(c, email)
	})
	gameIds := make([]*datastore.Key, len(result))
	for index, gameMember := range result {
		gameIds[index] = gameMember.GameId
		result[index].Id = result[index].IdByEmail(c)
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
	return &result
}

func GetGameMemberById(c appengine.Context, key *datastore.Key) *GameMember {
	var result GameMember
	if common.Memoize(c, gameMemberByIdKey(key), &result, func() interface{} {
		return findGameMemberById(c, key)
	}) {
		result.Id = result.IdByEmail(c)
		result.Game = GetGamesByIds(c, []*datastore.Key{result.GameId})[0]
		result.Phase = GetLatestPhasesByGameIds(c, []*datastore.Key{result.GameId})[0]
		return &result
	}
	return nil
}
