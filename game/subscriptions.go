package game

import (
	"encoding/base64"
	"fmt"
	"sort"
	"time"
	"code.google.com/p/go.net/websocket"

	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/wsubs/gosubs"
)

type MemberState struct {
	*Member
	User *user.User
}

type GameState struct {
	*Game
	Members        []MemberState
	UnseenMessages map[string]int
	TimeLeft       time.Duration
	Phase          *Phase
}

type GameStates []GameState

func (self GameStates) Len() int {
	return len(self)
}

func (self GameStates) Less(i, j int) bool {
	return self[j].Game.CreatedAt.Before(self[i].Game.CreatedAt)
}

func (self GameStates) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func SubscribeCurrent(c common.WSContext) error {
	if c.Principal() == "" {
		return websocket.JSON.Send(c.Conn(), gosubs.Message{
			Type: gosubs.FetchType,
			Object: &gosubs.Object{
				URI: c.Match()[0],
			},
		})
	}
	s := c.Pack().New(c.Match()[0])
	s.Query = s.DB().Query().Where(kol.Equals{"UserId", kol.Id(c.Principal())})
	s.Call = func(i interface{}, op string) error {
		members := i.([]*Member)
		states := GameStates{}
		for _, member := range members {
			if op == gosubs.DeleteType {
				states = append(states, GameState{
					Game:    &Game{Id: member.GameId},
					Members: []MemberState{MemberState{Member: member}},
				})
			} else {
				game := &Game{Id: member.GameId}
				if err := s.DB().Get(game); err != nil {
					return err
				}
				members, err := game.Members(c.DB())
				if err != nil {
					return err
				}
				state, err := game.ToState(c.DB(), members, member)
				if err != nil {
					return err
				}
				states = append(states, state)
			}
		}
		if op == gosubs.FetchType || len(states) > 0 {
			sort.Sort(states)
			return s.Send(states, op)
		}
		return nil
	}
	return s.Subscribe(&Member{})
}

func SubscribeGame(c common.WSContext) error {
	base64DecodedId, err := base64.URLEncoding.DecodeString(c.Match()[1])
	if err != nil {
		return err
	}
	s := c.Pack().New(c.Match()[0])
	s.Call = func(i interface{}, op string) error {
		game := i.(*Game)
		members, err := game.Members(c.DB())
		if err != nil {
			return err
		}
		member := members.Get(c.Principal())
		isMember := member != nil
		if !game.Private || isMember {
			state, err := game.ToState(c.DB(), members, member)
			if err != nil {
				return err
			}
			return s.Send(state, op)
		} else if op == gosubs.FetchType {
			return s.Send(GameState{}, op)
		}
		return nil
	}
	return s.Subscribe(&Game{Id: base64DecodedId})
}

func SubscribeMessages(c common.WSContext) (err error) {
	base64DecodedId, err := base64.URLEncoding.DecodeString(c.Match()[1])
	if err != nil {
		return err
	}
	game := &Game{Id: base64DecodedId}
	if err = c.DB().Get(game); err != nil {
		return
	}
	variant, found := common.VariantMap[game.Variant]
	if !found {
		err = fmt.Errorf("Unknown variant for %+v", game)
		return
	}
	member, err := game.Member(c.DB(), c.Principal())
	if err != nil && err != kol.NotFound {
		return
	}
	memberId := ""
	if member != nil {
		memberId = member.Id.String()
	}
	s := c.Pack().New(c.Match()[0])
	s.Query = s.DB().Query().Where(kol.Equals{"GameId", base64DecodedId})
	s.Call = func(i interface{}, op string) (err error) {
		messages := i.([]*Message)
		result := Messages{}
		for _, message := range messages {
			if len(message.RecipientIds) == len(variant.Nations) || message.RecipientIds[memberId] {
				result = append(result, *message)
			}
		}
		if op == gosubs.FetchType || len(result) > 0 {
			sort.Sort(result)
			return s.Send(result, op)
		}
		return
	}
	return s.Subscribe(&Message{})
}

func SubscribeOpen(c common.WSContext) error {
	if c.Principal() == "" {
		return websocket.JSON.Send(c.Conn(), gosubs.Message{
			Type: gosubs.FetchType,
			Object: &gosubs.Object{
				URI: c.Match()[0],
			},
		})
	}
	s := c.Pack().New(c.Match()[0])
	s.Query = s.DB().Query().Where(kol.And{kol.Equals{"Closed", false}, kol.Equals{"Private", false}})
	s.Call = func(i interface{}, op string) error {
		games := i.([]*Game)
		states := GameStates{}
		isMember := false
		me := &user.User{Id: kol.Id(c.Principal())}
		if err := c.DB().Get(me); err != nil {
			return err
		}
		for _, game := range games {
			if game.Disallows(me) {
				break
			}
			members, err := game.Members(c.DB())
			if err != nil {
				return err
			}
			if disallows, err := members.Disallows(c.DB(), me); err != nil {
				return err
			} else if disallows {
				break
			}
			isMember = members.Contains(c.Principal())
			if !isMember {
				state, err := game.ToState(c.DB(), members, nil)
				if err != nil {
					return err
				}
				states = append(states, state)
			}
		}
		if op == gosubs.FetchType || len(states) > 0 {
			sort.Sort(states)
			return s.Send(states, op)
		}
		return nil
	}
	return s.Subscribe(&Game{})
}
