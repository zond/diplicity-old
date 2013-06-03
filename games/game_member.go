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

func (self *GameMember) ValidatedDelete(c appengine.Context, email string) {
	if err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		// validate that user is member
		if self.Email != email {
			return fmt.Errorf("%+v is not %v", self, email)
		}
		var game *Game
		if game = GetGameById(c, self.Id.Parent(), false); game.Started {
			return fmt.Errorf("%+v has already started", game)
		}
		if err := datastore.Delete(c, self.Id); err != nil {
			return err
		}

		common.MemDel(c, gameMembersByUserKey(email), gameMembersByGameKey(self.Id.Parent()), gameMemberByIdKey(self.Id))

		otherMembers := false
		for _, memb := range game.GetMembers(c, false) {
			if memb.Email != self.Email {
				otherMembers = true
				break
			}
		}
		if !otherMembers {
			if err := game.Delete(c); err != nil {
				return err
			}
		}

		return nil
	}, &datastore.TransactionOptions{XG: false}); err != nil {
		panic(err)
	}
}

func (self *GameMember) SaveWithGame(c appengine.Context, email string) (result *GameMember, err error) {
	result = self
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		if self.Game, err = self.Game.save(c, email); err != nil {
			return err
		}
		_, err = self.save(c, email, self.Game.Id)
		return err
	}, &datastore.TransactionOptions{XG: false})
	self.memEnsure(c)
	return
}

func (self *GameMember) SaveWithoutGame(c appengine.Context, email string) (result *GameMember, err error) {
	result = self
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		_, err = self.save(c, email, self.Game.Id)
		return err
	}, &datastore.TransactionOptions{XG: false})
	self.memEnsure(c)
	return
}

func (self *GameMember) save(c appengine.Context, email string, gameId *datastore.Key) (result *GameMember, err error) {
	result = self

	members := make(map[string]*GameMember)
	for _, memb := range self.Game.GetMembers(c, false) {
		members[memb.Id.Encode()] = memb
	}
	if len(members) > len(common.VariantMap[self.Game.Variant].Nations) {
		err = fmt.Errorf("%+v would get too many members", game)
		return
	}

	if self.Id == nil {
		self.Id, err = datastore.Put(c, datastore.NewKey(c, gameMemberKind, self.Email, 0, gameId), self)
	} else {
		_, err = datastore.Put(c, datastore.NewKey(c, gameMemberKind, self.Email, 0, gameId), self)
	}
	if err != nil {
		return
	}
	members[self.Id.Encode()] = self

	if self.Game == nil {
		self.Game = GetGameById(c, gameId, false)
	}
	if !self.Game.Started && len(members) == len(common.VariantMap[self.Game.Variant].Nations) {
		if err = self.Game.start(c, members); err != nil {
			return
		}
	}

	common.MemDel(c, gameMembersByUserKey(email), gameMembersByGameKey(self.Id.Parent()), gameMemberByIdKey(self.Id))
	return
}

func (self *GameMember) memEnsure(c appengine.Context) {
	cpy := *self
	cpy.Game = nil
	cpy.Phase = nil
	currentMembers := GetGameMembersByUser(c, self.Email, true)
	found := false
	for index, _ := range currentMembers {
		if currentMembers[index].Id.Equal(cpy.Id) {
			currentMembers[index] = &cpy
			found = true
		}
	}
	if !found {
		currentMembers = append(currentMembers, &cpy)
	}
	common.MemPut(c, gameMembersByUserKey(self.Email), currentMembers)
}

func findGameMembersByGameId(c appengine.Context, gameId *datastore.Key) (result GameMembers) {
	ids, err := datastore.NewQuery(gameMemberKind).Ancestor(gameId).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		result[index].Id = id
	}
	return
}

func (self *Game) GetMembers(c appengine.Context, useCache bool) (result GameMembers) {
	if useCache {
		common.Memoize(c, gameMembersByGameKey(self.Id), &result, func() interface{} {
			return findGameMembersByGameId(c, self.Id)
		})
	} else {
		result = findGameMembersByGameId(c, self.Id)
	}
	for index, _ := range result {
		result[index].Game = self
	}
	if result == nil {
		result = make(GameMembers, 0)
	}
	return
}

func findGameMembersByUser(c appengine.Context, email string) (result GameMembers) {
	ids, err := datastore.NewQuery(gameMemberKind).Filter("Email=", email).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		result[index].Id = id
	}
	return
}

func GetGameMembersByUser(c appengine.Context, email string, useCache bool) (result GameMembers) {
	if useCache {
		common.Memoize(c, gameMembersByUserKey(email), &result, func() interface{} {
			return findGameMembersByUser(c, email)
		})
	} else {
		result = findGameMembersByUser(c, email)
	}
	gameIds := make([]*datastore.Key, len(result))
	for index, gameMember := range result {
		gameIds[index] = gameMember.Id.Parent()
	}
	for index, game := range GetGamesByIds(c, gameIds) {
		gameCopy := game
		result[index].Game = gameCopy
		if game.Owner == email {
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

func GetOpenGamesForUser(c appengine.Context, email string) (result GameMembers) {
	memberMap := make(map[string]bool)
	for _, member := range GetGameMembersByUser(c, email, true) {
		memberMap[member.Id.Parent().Encode()] = true
	}

	for _, game := range GetOpenGames(c) {
		if !memberMap[game.Id.Encode()] {
			owner := game.Owner == email
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
	if err == datastore.ErrNoSuchEntity {
		return nil
	}
	result.Id = key
	common.AssertOkError(err)
	return &result
}

func GetGameMemberById(c appengine.Context, key *datastore.Key) *GameMember {
	var result GameMember
	if common.Memoize(c, gameMemberByIdKey(key), &result, func() interface{} {
		return findGameMemberById(c, key)
	}) {
		result.Game = GetGameById(c, result.Id.Parent(), true)
		result.Phase = GetLatestPhasesByGameIds(c, []*datastore.Key{result.Id.Parent()})[0]
		return &result
	}
	return nil
}
