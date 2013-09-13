package game

import (
	"bytes"
	"github.com/zond/diplicity/common"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
	"regexp"
)

type Minutes int

var URIPattern = regexp.MustCompile("^/games/(.*)$")

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

func Update(d *kol.DB, m map[string]interface{}, updater string) {
	var state gameMemberState
	common.MustUnmarshalJSON(common.MustMarshalJSON(m), &state)

	game := &Game{Id: state.Game.Id}
	if err := d.Get(game); err != nil {
		panic(err)
	}
	if bytes.Compare(game.Owner, []byte(updater)) == 0 {
		game.Variant, game.EndYear, game.Private, game.Deadlines, game.ChatFlags =
			state.Game.Variant, state.Game.EndYear, state.Game.Private, state.Game.Deadlines, state.Game.ChatFlags
		if err := d.Set(game); err != nil {
			panic(err)
		}
	}
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

func SubscribeCurrent(s *subs.Subscription, email string) {
	s.Query = s.DB().Query().Where(kol.Equals{"User", []byte(email)})
	s.Call = func(i interface{}, op string) {
		members := i.([]*Member)
		states := []gameMemberState{}
		for _, member := range members {
			game := &Game{Id: member.Game}
			if err := s.DB().Get(game); err != nil {
				panic(err)
			}
			if !game.Ended {
				phase := &Phase{Id: game.Phase}
				if err := s.DB().Get(phase); err != nil {
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
		s.Send(states, op)
	}
	s.Subscribe(new(Member))
}

func SubscribeOpen(s *subs.Subscription, email string) {
	s.Query = s.DB().Query().Where(kol.And{kol.Equals{"Closed", false}, kol.Equals{"Private", false}})
	s.Call = func(i interface{}, op string) {
		var members []Member
		games := i.([]*Game)
		states := []gameMemberState{}
		for _, game := range games {
			members = nil
			s.DB().Query().Where(kol.And{kol.Equals{"User", []byte(email)}, kol.Equals{"Game", game.Id}}).All(&members)
			if len(members) == 0 {
				phase := &Phase{Id: game.Phase}
				if err := s.DB().Get(phase); err != nil {
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
		s.Send(states, op)
	}
	s.Subscribe(new(Game))
}
