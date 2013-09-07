package game

import (
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/db"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
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

	Deadlines map[dip.PhaseType]Minutes

	ChatFlags map[dip.PhaseType]common.ChatFlag
}

func Create(m map[string]interface{}, owner interface{}) {
	g := &Game{
		Owner:   []byte(owner.(string)),
		Variant: m["Variant"].(string),
		EndYear: m["EndYear"].(int),
		Private: m["Private"].(bool),
	}
	fmt.Println("want to create %+v with %+v", g, m)
}

func (self *Game) Updated(d *kol.DB, old *Game) {
	members := []Member{}
	if err := d.Query().Where(kol.Equals{"Game", self.Id}).All(&members); err != nil {
		panic(err)
	}
	for _, member := range members {
		d.EmitUpdate(&member)
	}
}

type Phase struct {
	Id   []byte
	Game []byte `kol:"index"`

	Season  dip.Season
	Year    int
	Type    dip.PhaseType
	Ordinal int
}

func (self *Phase) Updated(d *kol.DB, old *Phase) {
	g := Game{Id: self.Game}
	if err := d.Get(&g); err != nil {
		panic(err)
	}
	d.EmitUpdate(&g)
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

func SubscribeCurrent(s *subs.Subscription, email interface{}) {
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
		db.SubscribeQuery(s.Name(), refinery, db.DB.Query().Where(kol.Equals{"User", []byte(email.(string))}), new(Member))
	}
}

func SubscribeOpen(s *subs.Subscription, email interface{}) {
	if email != nil {
		refinery := func(i interface{}, op string) {
			var members []Member
			games := i.([]*Game)
			states := []gameMemberState{}
			for _, game := range games {
				members = nil
				db.DB.Query().Where(kol.And{kol.Equals{"User", []byte(email.(string))}, kol.Equals{"Game", game.Id}}).All(&members)
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
		db.SubscribeQuery(s.Name(), refinery, db.DB.Query().Where(kol.And{kol.Equals{"Closed", false}, kol.Equals{"Private", false}}), new(Game))
	}
}
