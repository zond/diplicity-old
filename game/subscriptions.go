package game

import (
	"encoding/base64"
	"sort"
	"strconv"
	"time"
	"github.com/zond/diplicity/epoch"
	"github.com/zond/diplicity/game/meta"
	"github.com/zond/diplicity/srv"
	"github.com/zond/diplicity/user"
	"github.com/zond/godip/variants"
	"github.com/zond/unbolted"
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

func SubscribeMine(c srv.WSContext) error {
	if c.Principal() == "" {
		return c.Conn().WriteJSON(gosubs.Message{
			Type: gosubs.FetchType,
			Object: &gosubs.Object{
				URI: c.Match()[0],
			},
		})
	}
	s := c.Pack().New(c.Match()[0])
	s.Query = s.DB().Query().Where(unbolted.Equals{"UserId", unbolted.Id(c.Principal())})
	s.Call = func(i interface{}, op string) (err error) {
		return c.View(func(c srv.WSTXContext) (err error) {
			members := i.([]*Member)
			var ep time.Duration
			ep, err = epoch.Get(c.TX())
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
					if gameMembers, err = game.Members(c.TX()); err != nil {
						return
					}
					var state GameState
					if state, err = game.ToState(c.TX(), gameMembers, member); err != nil {
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
					case meta.GameStateStarted:
						_, phase, err := a.Game.Phase(c.TX(), 0)
						if err == nil {
							urgencyA = phase.Deadline - ep
						}
					case meta.GameStateCreated:
						urgencyA -= 1
					}
					switch b.State {
					case meta.GameStateStarted:
						_, phase, err := b.Game.Phase(c.TX(), 0)
						if err == nil {
							urgencyB = phase.Deadline - ep
						}
					case meta.GameStateCreated:
						urgencyB -= 1
					}
					if urgencyA != urgencyB {
						return urgencyA < urgencyB
					}
					return a.CreatedAt.Before(b.CreatedAt)
				}, 1024*16)
				return s.Send(states, op)
			}
			return
		})
	}
	return s.Subscribe(&Member{})
}

func SubscribeGame(c srv.WSContext) error {
	base64DecodedId, err := base64.URLEncoding.DecodeString(c.Match()[1])
	if err != nil {
		return err
	}
	s := c.Pack().New(c.Match()[0])
	s.Call = func(i interface{}, op string) (err error) {
		return c.View(func(c srv.WSTXContext) (err error) {
			game := i.(*Game)
			members, err := game.Members(c.TX())
			if err != nil {
				return err
			}
			member := members.Get(c.Principal())
			isMember := member != nil
			if !game.Private || isMember {
				state, err := game.ToState(c.TX(), members, member)
				if err != nil {
					return err
				}
				return s.Send(state, op)
			} else if op == gosubs.FetchType {
				return s.Send(GameState{}, op)
			}
			return
		})
	}
	return s.Subscribe(&Game{Id: base64DecodedId})
}

func SubscribeGamePhase(c srv.WSContext) (err error) {
	base64DecodedId, err := base64.URLEncoding.DecodeString(c.Match()[1])
	if err != nil {
		return err
	}
	if err = c.View(func(c srv.WSTXContext) (err error) {
		game := &Game{Id: base64DecodedId}
		if err = c.TX().Get(game); err != nil {
			return
		}
		phaseOrdinal, err := strconv.Atoi(c.Match()[2])
		if err != nil {
			return
		}
		members, err := game.Members(c.TX())
		if err != nil {
			return
		}
		member := members.Get(c.Principal())
		isMember := member != nil
		if !game.Private || isMember {
			state, err := game.ToStateWithPhaseOrdinal(c.TX(), members, member, phaseOrdinal)
			if err != nil {
				return err
			}
			return c.Conn().WriteJSON(gosubs.Message{
				Type: gosubs.FetchType,
				Object: &gosubs.Object{
					URI:  c.Match()[0],
					Data: state,
				},
			})
		}
		return
	}); err != nil {
		return
	}
	return
}

func SubscribeMessages(c srv.WSContext) (err error) {
	base64DecodedId, err := base64.URLEncoding.DecodeString(c.Match()[1])
	if err != nil {
		return err
	}
	game := &Game{Id: base64DecodedId}
	var member *Member
	if err = c.View(func(c srv.WSTXContext) (err error) {
		if err = c.TX().Get(game); err != nil {
			return
		}
		if member, err = game.Member(c.TX(), c.Principal()); err != nil {
			if err != unbolted.ErrNotFound {
				return
			}
			err = nil
		}
		return
	}); err != nil {
		return
	}
	memberId := ""
	if member != nil {
		memberId = member.Id.String()
	}
	s := c.Pack().New(c.Match()[0])
	s.Query = s.DB().Query().Where(unbolted.Equals{"GameId", base64DecodedId})
	s.Call = func(i interface{}, op string) (err error) {
		return c.View(func(c srv.WSTXContext) (err error) {
			messages := i.([]*Message)
			result := Messages{}
			for _, message := range messages {
				if message.Public {
					game := &Game{Id: message.GameId}
					if err = c.TX().Get(game); err != nil {
						return
					}
					var members Members
					if members, err = game.Members(c.TX()); err != nil {
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
		})
	}
	return s.Subscribe(&Message{})
}

func subscribeOthers(c srv.WSContext, filter unbolted.QFilter, preLimiter func(source Games) (result Games), postLimiter func(source GameStates) (result GameStates)) error {
	s := c.Pack().New(c.Match()[0])
	s.Query = s.DB().Query().Where(filter)
	s.Call = func(i interface{}, op string) (err error) {
		return c.View(func(c srv.WSTXContext) (err error) {
			games := i.([]*Game)
			if preLimiter != nil {
				games = ([]*Game)(preLimiter(Games(games)))
			}
			states := GameStates{}
			isMember := false
			me := &user.User{Id: unbolted.Id(c.Principal())}
			if err = c.TX().Get(me); err != nil {
				if err == unbolted.ErrNotFound {
					me = nil
					err = nil
				} else {
					return
				}
			}
			for _, game := range games {
				if !game.Disallows(me) {
					members := Members{}
					if members, err = game.Members(c.TX()); err != nil {
						return
					}
					disallows := false
					if disallows, err = members.Disallows(c.TX(), me); err != nil {
						return
					} else if !disallows {
						isMember = members.Contains(c.Principal())
						if !isMember {
							var state GameState
							if state, err = game.ToState(c.TX(), members, nil); err != nil {
								return
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
			return
		})
	}
	return s.Subscribe(&Game{})
}

func SubscribeOthersOpen(c srv.WSContext) (err error) {
	return subscribeOthers(c, unbolted.And{unbolted.Equals{"Closed", false}, unbolted.Equals{"Private", false}}, nil, func(source GameStates) (result GameStates) {
		return source.SortAndLimit(func(a, b GameState) bool {
			leftA := 0
			leftB := 0
			if variant, found := variants.Variants[a.Variant]; found {
				leftA = len(variant.Nations) - len(a.Members)
			}
			if variant, found := variants.Variants[b.Variant]; found {
				leftB = len(variant.Nations) - len(b.Members)
			}
			if leftA != leftB {
				return leftA < leftB
			}
			return a.CreatedAt.Before(b.CreatedAt)
		}, 128)
	})
}

func SubscribeOthersClosed(c srv.WSContext) error {
	return subscribeOthers(c, unbolted.And{unbolted.Equals{"State", meta.GameStateStarted}, unbolted.Equals{"Closed", true}, unbolted.Equals{"Private", false}}, func(source Games) (result Games) {
		return source.SortAndLimit(func(a, b *Game) bool {
			return a.UpdatedAt.Before(b.UpdatedAt)
		}, 128)
	}, nil)
}

func SubscribeOthersFinished(c srv.WSContext) error {
	return subscribeOthers(c, unbolted.And{unbolted.Equals{"State", meta.GameStateEnded}, unbolted.Equals{"Private", false}}, func(source Games) (result Games) {
		return source.SortAndLimit(func(a, b *Game) bool {
			return a.UpdatedAt.Before(b.UpdatedAt)
		}, 128)
	}, nil)
}
