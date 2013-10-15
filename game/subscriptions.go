package game

import (
	"encoding/base64"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
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

func SubscribeCurrent(c common.Context, s *subs.Subscription, email string) error {
	s.Query = s.DB().Query().Where(kol.Equals{"UserId", []byte(email)})
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
		if len(states) > 0 {
			return s.Send(states, op)
		}
		return nil
	}
	return s.Subscribe(new(Member))
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
		}
		return nil
	}
	return s.Subscribe(&Game{Id: base64DecodedId})
}

func SubscribeOpen(c common.Context, s *subs.Subscription, email string) error {
	s.Query = s.DB().Query().Where(kol.And{kol.Equals{"Closed", false}, kol.Equals{"Private", false}})
	s.Call = func(i interface{}, op string) error {
		games := i.([]*Game)
		states := []GameState{}
		isMember := false
		me := &user.User{Id: []byte(email)}
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
		if len(states) > 0 {
			return s.Send(states, op)
		}
		return nil
	}
	return s.Subscribe(new(Game))
}
