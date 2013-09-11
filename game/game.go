package game

import (
	"fmt"
	"github.com/zond/diplicity/common"
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

func Create(d *kol.DB, m map[string]interface{}, owner string) {
	var state gameMemberState
	common.MustUnmarshalJSON(common.MustMarshalJSON(m), &state)

	game := &Game{
		Owner:     []byte(owner),
		Variant:   state.Game.Variant,
		EndYear:   state.Game.EndYear,
		Private:   state.Game.Private,
		Deadlines: state.Game.Deadlines,
		ChatFlags: state.Game.ChatFlags,
	}

	member := &Member{
		User: []byte(owner),
	}
	d.Transact(func(d *kol.DB) error {
		if err := d.Set(game); err != nil {
			return err
		}
		member.Game = game.Id
		return d.Set(member)
	})
	fmt.Printf("created %+v\nfrom%v\n", game, m)
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
	*Member
	Game  *Game
	Phase *Phase
}

func CurrentSubscription(db *kol.DB, s *subs.Subscription, email string) *common.Subscription {
	return &common.Subscription{
		Name: s.Name(),
		Subscriber: func(i interface{}, op string) {
			members := i.([]*Member)
			states := []gameMemberState{}
			for _, member := range members {
				game := &Game{Id: member.Game}
				if err := db.Get(game); err != nil {
					panic(err)
				}
				if !game.Ended {
					phase := &Phase{Id: game.Phase}
					if err := db.Get(phase); err != nil {
						if err == kol.NotFound {
							phase = nil
						} else {
							panic(err)
						}
					}
					states = append(states, gameMemberState{
						Member: member,
						Game:   game,
						Phase:  phase,
					})
				}
			}
			s.Call(states, op)
		},
		Query:  db.Query().Where(kol.Equals{"User", []byte(email)}),
		Object: new(Member),
	}
}

func OpenSubscription(db *kol.DB, s *subs.Subscription, email string) *common.Subscription {
	return &common.Subscription{
		Name: s.Name(),
		Subscriber: func(i interface{}, op string) {
			var members []Member
			games := i.([]*Game)
			states := []gameMemberState{}
			for _, game := range games {
				members = nil
				db.Query().Where(kol.And{kol.Equals{"User", []byte(email)}, kol.Equals{"Game", game.Id}}).All(&members)
				if len(members) == 0 {
					phase := &Phase{Id: game.Phase}
					if err := db.Get(phase); err != nil {
						if err == kol.NotFound {
							phase = nil
						} else {
							panic(err)
						}
					}
					states = append(states, gameMemberState{
						Game:  game,
						Phase: phase,
					})
				}
			}
			s.Call(states, op)
		},
		Query:  db.Query().Where(kol.And{kol.Equals{"Closed", false}, kol.Equals{"Private", false}}),
		Object: new(Game),
	}
}
