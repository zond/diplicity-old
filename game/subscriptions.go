package game

import (
	"encoding/base64"
	"sort"
	"strconv"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/epoch"
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
	Phases         int
}

type GameStates []GameState

func (self GameStates) SortAndLimit(f func(a, b GameState) bool, limit int) GameStates {
	sorted := SortedGameStates{
		GameStates: self,
		LessFunc:   f,
	}
	sort.Sort(sorted)
	if len(sorted.GameStates) > limit {
		return sorted.GameStates[:limit]
	}
	return sorted.GameStates
}

type SortedGameStates struct {
	GameStates GameStates
	LessFunc   func(a, b GameState) bool
}

func (self SortedGameStates) Len() int {
	return len(self.GameStates)
}

func (self SortedGameStates) Less(i, j int) bool {
	return self.LessFunc(self.GameStates[i], self.GameStates[j])
}

func (self SortedGameStates) Swap(i, j int) {
	self.GameStates[j], self.GameStates[i] = self.GameStates[i], self.GameStates[j]
}

func SubscribeMine(c common.WSContext) error {
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
	s.Call = func(i interface{}, op string) (err error) {
		members := i.([]*Member)
		var ep time.Duration
		ep, err = epoch.Get(c.DB())
		if err != nil {
			return
		}
		states := GameStates{}
		for _, member := range members {
			if op == gosubs.DeleteType {
				states = append(states, GameState{
					Game:    &Game{Id: member.GameId},
					Members: []MemberState{MemberState{Member: member}},
				})
			} else {
				game := &Game{Id: member.GameId}
				if err = s.DB().Get(game); err != nil {
					return
				}
				var gameMembers Members
				if gameMembers, err = game.Members(c.DB()); err != nil {
					return
				}
				var state GameState
				if state, err = game.ToState(c.DB(), gameMembers, member); err != nil {
					return
				}
				states = append(states, state)
			}
		}
		if op == gosubs.FetchType || len(states) > 0 {
			states = states.SortAndLimit(func(a, b GameState) bool {
				urgencyA := time.Hour * 24 * 365
				urgencyB := time.Hour * 24 * 365
				switch a.State {
				case common.GameStateStarted:
					_, phase, err := a.Game.Phase(c.DB(), 0)
					if err == nil {
						urgencyA = phase.Deadline - ep
					}
				case common.GameStateCreated:
					urgencyA -= 1
				}
				switch b.State {
				case common.GameStateStarted:
					_, phase, err := b.Game.Phase(c.DB(), 0)
					if err == nil {
						urgencyB = phase.Deadline - ep
					}
				case common.GameStateCreated:
					urgencyB -= 1
				}
				if urgencyA != urgencyB {
					return urgencyA < urgencyB
				}
				return a.CreatedAt.Before(b.CreatedAt)
			}, 1024*16)
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

func SubscribeGamePhase(c common.WSContext) error {
	base64DecodedId, err := base64.URLEncoding.DecodeString(c.Match()[1])
	if err != nil {
		return err
	}
	game := &Game{Id: base64DecodedId}
	if err = c.DB().Get(game); err != nil {
		return err
	}
	phaseOrdinal, err := strconv.Atoi(c.Match()[2])
	if err != nil {
		return err
	}
	members, err := game.Members(c.DB())
	if err != nil {
		return err
	}
	member := members.Get(c.Principal())
	isMember := member != nil
	if !game.Private || isMember {
		state, err := game.ToStateWithPhaseOrdinal(c.DB(), members, member, phaseOrdinal)
		if err != nil {
			return err
		}
		return websocket.JSON.Send(c.Conn(), gosubs.Message{
			Type: gosubs.FetchType,
			Object: &gosubs.Object{
				URI:  c.Match()[0],
				Data: state,
			},
		})
	}
	return nil
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
			if message.Public {
				game := &Game{Id: message.GameId}
				if err = c.DB().Get(game); err != nil {
					return
				}
				var members Members
				if members, err = game.Members(c.DB()); err != nil {
					return
				}
				message.RecipientIds = map[string]bool{}
				for _, memb := range members {
					message.RecipientIds[memb.Id.String()] = true
				}
			}
			if message.Public || message.RecipientIds[memberId] {
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

func subscribeOthers(c common.WSContext, filter kol.QFilter, preLimiter func(source Games) (result Games), postLimiter func(source GameStates) (result GameStates)) error {
	if c.Principal() == "" {
		return websocket.JSON.Send(c.Conn(), gosubs.Message{
			Type: gosubs.FetchType,
			Object: &gosubs.Object{
				URI: c.Match()[0],
			},
		})
	}
	s := c.Pack().New(c.Match()[0])
	s.Query = s.DB().Query().Where(filter)
	s.Call = func(i interface{}, op string) error {
		games := i.([]*Game)
		if preLimiter != nil {
			games = ([]*Game)(preLimiter(Games(games)))
		}
		states := GameStates{}
		isMember := false
		me := &user.User{Id: kol.Id(c.Principal())}
		if err := c.DB().Get(me); err != nil {
			return err
		}
		for _, game := range games {
			if !game.Disallows(me) {
				members, err := game.Members(c.DB())
				if err != nil {
					return err
				}
				if disallows, err := members.Disallows(c.DB(), me); err != nil {
					return err
				} else if !disallows {
					isMember = members.Contains(c.Principal())
					if !isMember {
						state, err := game.ToState(c.DB(), members, nil)
						if err != nil {
							return err
						}
						states = append(states, state)
					}
				}
			}
		}
		if op == gosubs.FetchType || len(states) > 0 {
			if postLimiter != nil {
				states = postLimiter(states)
			}
			return s.Send(states, op)
		}
		return nil
	}
	return s.Subscribe(&Game{})
}

func SubscribeOthersOpen(c common.WSContext) error {
	return subscribeOthers(c, kol.And{kol.Equals{"Closed", false}, kol.Equals{"Private", false}}, nil, func(source GameStates) (result GameStates) {
		return source.SortAndLimit(func(a, b GameState) bool {
			leftA := 0
			leftB := 0
			if variant, found := common.Variants[a.Variant]; found {
				leftA = len(variant.Nations) - len(a.Members)
			}
			if variant, found := common.Variants[b.Variant]; found {
				leftB = len(variant.Nations) - len(b.Members)
			}
			if leftA != leftB {
				return leftA < leftB
			}
			return a.CreatedAt.Before(b.CreatedAt)
		}, 128)
	})
}

func SubscribeOthersClosed(c common.WSContext) error {
	return subscribeOthers(c, kol.And{kol.Equals{"State", common.GameStateStarted}, kol.Equals{"Closed", true}, kol.Equals{"Private", false}}, func(source Games) (result Games) {
		return source.SortAndLimit(func(a, b *Game) bool {
			return a.UpdatedAt.Before(b.UpdatedAt)
		}, 128)
	}, nil)
}

func SubscribeOthersFinished(c common.WSContext) error {
	return subscribeOthers(c, kol.And{kol.Equals{"State", common.GameStateEnded}, kol.Equals{"Private", false}}, func(source Games) (result Games) {
		return source.SortAndLimit(func(a, b *Game) bool {
			return a.UpdatedAt.Before(b.UpdatedAt)
		}, 128)
	}, nil)
}
