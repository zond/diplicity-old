package game

import (
	"encoding/base64"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
	"sort"
)

type MemberState struct {
	*Member
	User *user.User
}

type GameState struct {
	*Game
	Members []MemberState
	Phase   *Phase
}

type MessageState struct {
	*Message
	Member *Member
}

func SubscribeCurrent(c common.Context, s *subs.Subscription, email string) error {
	s.Query = s.DB().Query().Where(kol.Equals{"UserId", kol.Id(email)})
	s.Call = func(i interface{}, op string) error {
		members := i.([]*Member)
		states := []GameState{}
		for _, member := range members {
			if op == common.DeleteType {
				states = append(states, GameState{
					Game:    &Game{Id: member.GameId},
					Members: []MemberState{MemberState{Member: member}},
				})
			} else {
				game := &Game{Id: member.GameId}
				if err := s.DB().Get(game); err != nil {
					return err
				}
				if !game.Ended {
					states = append(states, GameState{
						Game:    game,
						Members: game.Members(c.DB()).toStates(c, game, email),
						Phase:   game.LastPhase(c.DB()),
					})
				}
			}
		}
		if op == subs.FetchType || len(states) > 0 {
			return s.Send(states, op)
		}
		return nil
	}
	return s.Subscribe(&Member{})
}

func SubscribeGame(c common.Context, s *subs.Subscription, gameId, email string) error {
	base64DecodedId, err := base64.URLEncoding.DecodeString(gameId)
	if err != nil {
		return err
	}
	s.Call = func(i interface{}, op string) error {
		game := i.(*Game)
		members := game.Members(c.DB())
		isMember := false
		for _, m := range members {
			if string(m.UserId) == email {
				isMember = true
				break
			}
		}
		if !game.Private || isMember {
			return s.Send(GameState{
				Game:    game,
				Members: members.toStates(c, game, email),
				Phase:   game.LastPhase(c.DB()),
			}, op)
		} else if op == subs.FetchType {
			return s.Send(GameState{}, op)
		}
		return nil
	}
	return s.Subscribe(&Game{Id: base64DecodedId})
}

type messagePointers []*Message

func (self messagePointers) Len() int {
	return len(self)
}

func (self messagePointers) Less(j, i int) bool {
	return self[i].CreatedAt.Before(self[j].CreatedAt)
}

func (self messagePointers) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func SubscribeMessages(c common.Context, s *subs.Subscription, gameId, email string) error {
	base64DecodedId, err := base64.URLEncoding.DecodeString(gameId)
	if err != nil {
		return err
	}
	s.Query = s.DB().Query().Where(kol.Equals{"GameId", base64DecodedId})
	s.Call = func(i interface{}, op string) error {
		messages := i.([]*Message)
		sort.Sort(messagePointers(messages))
		if len(messages) > 200 {
			messages = messages[:200]
		}
		states := []MessageState{}
		for _, message := range messages {
			game := &Game{Id: base64DecodedId}
			if err := c.DB().Get(game); err != nil {
				return err
			}
			member, err := game.Member(s.DB(), email)
			if err != nil {
				return err
			}
			phase := game.LastPhase(c.DB())
			if game.MessageAllowed(phase, member, message) {
				state, err := message.toState(c.DB())
				if err != nil {
					states = append(states, *state)
				}
			}
		}
		if op == subs.FetchType || len(states) > 0 {
			return s.Send(states, op)
		}
		return nil
	}
	return s.Subscribe(&Message{})
}

func SubscribeOpen(c common.Context, s *subs.Subscription, email string) error {
	s.Query = s.DB().Query().Where(kol.And{kol.Equals{"Closed", false}, kol.Equals{"Private", false}})
	s.Call = func(i interface{}, op string) error {
		games := i.([]*Game)
		states := []GameState{}
		isMember := false
		me := &user.User{Id: kol.Id(email)}
		if err := c.DB().Get(me); err != nil {
			return err
		}
		for _, game := range games {
			if game.Disallows(me) {
				break
			}
			members := game.Members(c.DB())
			if disallows, err := members.Disallows(c.DB(), me); err != nil {
				return err
			} else if disallows {
				break
			}
			isMember = false
			for _, m := range members {
				if string(m.UserId) == email {
					isMember = true
					break
				}
			}
			if !isMember {
				states = append(states, GameState{
					Game:    game,
					Members: members.toStates(c, game, email),
					Phase:   game.LastPhase(c.DB()),
				})
			}
		}
		if op == subs.FetchType || len(states) > 0 {
			return s.Send(states, op)
		}
		return nil
	}
	return s.Subscribe(&Game{})
}
