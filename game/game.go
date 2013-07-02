package game

import (
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/db"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
)

type Minutes int

type Game struct {
	Id    []byte
	Owner []byte
	Phase []byte

	Closed  bool `kol:"index"`
	Started bool `kol:"index"`
	Ended   bool `kol:"index"`
	Variant string
	EndYear int
	Private bool `kol:"index"`

	Deadlines map[string]Minutes

	ChatFlags map[string]common.ChatFlag
}

type Phase struct {
	Id   []byte
	Game []byte `kol:"index"`

	Season  dip.Season
	Year    int
	Type    dip.PhaseType
	Ordinal int
}

type Member struct {
	Id   []byte
	User []byte `kol:"index"`
	Game []byte `kol:"index"`

	Nation dip.Nation
}

type gameMemberState struct {
	Id []byte
	*Game
	*Member
	*Phase
}

func SubscribeCurrent(s *db.Subscription, email interface{}) {
	if email != nil {
		refinery := func(i interface{}, op string) {
			members := i.([]*Member)
			states := []gameMemberState{}
			for _, member := range members {
				game := &Game{Id: member.Game}
				if err := db.DB.Get(game); err != nil {
					panic(err)
				}
				if !game.Ended {
					phase := &Phase{Id: game.Phase}
					if err := db.DB.Get(phase); err != nil {
						if err == kol.NotFound {
							phase = nil
						} else {
							panic(err)
						}
					}
					states = append(states, gameMemberState{
						Id:     game.Id,
						Game:   game,
						Phase:  phase,
						Member: member,
					})
				}
			}
			s.Call(states, op)
		}
		db.SubscribeQuery(s.Name(), refinery, common.DB.Query().Filter(kol.Equals{"User", []byte(email.(string))}), new(Member))
		s.Register()
	}
}

func SubscribeOpen(s *db.Subscription, email interface{}) {
	if email != nil {
		refinery := func(i interface{}, op string) {
			var members []Member
			games := i.([]*Game)
			states := []gameMemberState{}
			for _, game := range games {
				members = nil
				db.DB.Query().Filter(kol.And{kol.Equals{"User", []byte(email.(string))}, kol.Equals{"Game", game.Id}}).All(&members)
				if len(members) == 0 {
					phase := &Phase{Id: game.Phase}
					if err := db.DB.Get(phase); err != nil {
						if err == kol.NotFound {
							phase = nil
						} else {
							panic(err)
						}
					}
					states = append(states, gameMemberState{
						Id:    game.Id,
						Game:  game,
						Phase: phase,
					})
				}
			}
			s.Call(states, op)
		}
		db.SubscribeQuery(s.Name(), refinery, common.DB.Query().Filter(kol.Equals{"Closed", false}), new(Game))
		s.Register()
	}
}
