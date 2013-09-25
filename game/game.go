package game

import (
	"encoding/base64"
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	"github.com/zond/godip/classical"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/state"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
	"math/rand"
	"net/url"
	"sort"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Minutes int

type Game struct {
	Id []byte

	Closed           bool `kol:"index"`
	Started          bool `kol:"index"`
	Ended            bool `kol:"index"`
	Variant          string
	AllocationMethod string
	SecretEmail      bool
	SecretNickname   bool
	SecretNation     bool
	EndYear          int
	Private          bool `kol:"index"`

	Deadlines map[dip.PhaseType]Minutes

	ChatFlags map[dip.PhaseType]common.ChatFlag
}

func (self *Game) allocate(d *kol.DB) error {
	members := Members{}
	if err := d.Query().Where(kol.Equals{"GameId", self.Id}).All(&members); err != nil {
		return err
	}
	switch self.AllocationMethod {
	case common.RandomString:
		for memberIndex, nationIndex := range rand.Perm(len(members)) {
			members[memberIndex].Nation = common.VariantMap[self.Variant].Nations[nationIndex]
			if err := d.Set(&members[memberIndex]); err != nil {
				return err
			}
		}
		return nil
	case common.PreferencesString:
		prefs := make([][]dip.Nation, len(members))
		for index, member := range members {
			prefs[index] = member.PreferredNations
		}
		for index, nation := range optimizePreferences(prefs) {
			members[index].Nation = nation
			if err := d.Set(&members[index]); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("Unknown allocation method %v", self.AllocationMethod)
}

func (self *Game) start(d *kol.DB) error {
	self.Started = true
	self.Closed = true
	if err := d.Set(self); err != nil {
		return err
	}
	if err := self.allocate(d); err != nil {
		return err
	}
	var startState *state.State
	if self.Variant == common.ClassicalString {
		startState = classical.Start()
	} else {
		return fmt.Errorf("Unknown variant %v", self.Variant)
	}
	startPhase := startState.Phase()
	phase := &Phase{
		GameId:        self.Id,
		Ordinal:       0,
		Season:        startPhase.Season(),
		Year:          startPhase.Year(),
		Type:          startPhase.Type(),
		Units:         startState.Units(),
		SupplyCenters: startState.SupplyCenters(),
	}
	return d.Set(phase)
}

func DeleteMember(c common.Context, gameId, email string) {
	if err := c.DB().Transact(func(d *kol.DB) error {
		urlDecodedId, err := url.QueryUnescape(gameId)
		if err != nil {
			return err
		}
		base64DecodedId, err := base64.StdEncoding.DecodeString(urlDecodedId)
		if err != nil {
			return err
		}
		game := Game{Id: base64DecodedId}
		if err := d.Get(&game); err != nil {
			return fmt.Errorf("Game not found: %v", err)
		}
		member := Member{}
		if _, err := d.Query().Where(kol.And{kol.Equals{"GameId", base64DecodedId}, kol.Equals{"UserId", []byte(email)}}).First(&member); err != nil {
			return err
		}
		if !game.Started {
			if err := d.Del(&member); err != nil {
				return err
			}
			left := []Member{}
			if err := d.Query().Where(kol.Equals{"GameId", game.Id}).All(&left); err != nil {
				return err
			}
			if len(left) == 0 {
				if err := d.Del(&game); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		c.Errorf("Unable to delete member: %v", err)
	}
}

func AddMember(c common.Context, j common.JSON, email string) {
	var state GameState
	j.Overwrite(&state)
	if err := c.DB().Transact(func(d *kol.DB) error {
		game := Game{Id: state.Game.Id}
		if err := d.Get(&game); err != nil {
			return err
		}
		variant, found := common.VariantMap[game.Variant]
		if !found {
			return fmt.Errorf("Unknown variant %v", game.Variant)
		}
		already := []Member{}
		if err := d.Query().Where(kol.Equals{"GameId", state.Game.Id}).All(&already); err != nil {
			return err
		}
		if len(already) < len(variant.Nations) {
			id := make([]byte, len(state.Game.Id)+len([]byte(email)))
			copy(id, state.Game.Id)
			copy(id[len(state.Game.Id):], []byte(email))
			member := Member{
				Id:               id,
				GameId:           state.Game.Id,
				UserId:           []byte(email),
				PreferredNations: state.Members[0].PreferredNations,
			}
			if err := d.Set(&member); err != nil {
				return err
			}
			if len(already) == len(variant.Nations)-1 {
				if err := game.start(d); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		c.Errorf("Unable to add member: %v", err)
	}
}

func Create(c common.Context, j common.JSON, creator string) {
	var state GameState
	j.Overwrite(&state)

	game := &Game{
		Variant:          state.Game.Variant,
		EndYear:          state.Game.EndYear,
		Private:          state.Game.Private,
		SecretEmail:      state.Game.SecretEmail,
		SecretNickname:   state.Game.SecretNickname,
		SecretNation:     state.Game.SecretNation,
		Deadlines:        state.Game.Deadlines,
		ChatFlags:        state.Game.ChatFlags,
		AllocationMethod: state.Game.AllocationMethod,
	}

	member := &Member{
		UserId:           []byte(creator),
		PreferredNations: state.Members[0].PreferredNations,
	}
	c.DB().Transact(func(d *kol.DB) error {
		if err := d.Set(game); err != nil {
			return err
		}
		member.GameId = game.Id
		return d.Set(member)
	})
}

func (self *Game) Updated(d *kol.DB, old *Game) {
	members := []Member{}
	if err := d.Query().Where(kol.Equals{"GameId", self.Id}).All(&members); err != nil {
		panic(err)
	}
	for _, member := range members {
		d.EmitUpdate(&member)
	}
}

type Phase struct {
	Id     []byte
	GameId []byte `kol:"index"`

	Season  dip.Season
	Year    int
	Type    dip.PhaseType
	Ordinal int

	Units         map[dip.Province]dip.Unit
	SupplyCenters map[dip.Province]dip.Nation
}

func (self *Phase) Updated(d *kol.DB, old *Phase) {
	g := Game{Id: self.GameId}
	if err := d.Get(&g); err != nil {
		panic(err)
	}
	d.EmitUpdate(&g)
}

type Phases []Phase

func (self Phases) Len() int {
	return len(self)
}

func (self Phases) Less(i, j int) bool {
	return self[i].Ordinal < self[j].Ordinal
}

func (self Phases) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type Member struct {
	Id     []byte
	UserId []byte `kol:"index"`
	GameId []byte `kol:"index"`

	Nation           dip.Nation
	PreferredNations []dip.Nation
}

func (self *Member) Deleted(d *kol.DB) {
	g := Game{Id: self.GameId}
	if err := d.Get(&g); err == nil {
		d.EmitUpdate(&g)
	} else if err != kol.NotFound {
		panic(err)
	}
}

func (self *Member) Created(d *kol.DB) {
	g := Game{Id: self.GameId}
	if err := d.Get(&g); err != nil {
		panic(err)
	}
	d.EmitUpdate(&g)
}

type Members []Member

func (self Members) toStates(c common.Context, g *Game, email string) (result []MemberState) {
	result = make([]MemberState, len(self))
	for index, member := range self {
		cpy := member
		if string(cpy.UserId) != email {
			cpy.UserId = nil
			cpy.PreferredNations = nil
		}
		result[index] = MemberState{
			Member: &cpy,
			User:   &user.User{},
		}
		if !g.SecretEmail || !g.SecretNickname {
			foundUser := user.EnsureUser(c, string(member.UserId))
			if !g.SecretEmail {
				result[index].User.Email = foundUser.Email
				result[index].User.Id = foundUser.Id
			}
			if !g.SecretNickname {
				result[index].User.Nickname = foundUser.Nickname
			}
		}
	}
	return
}

type MemberState struct {
	*Member
	User *user.User
}

type GameState struct {
	*Game
	Members []MemberState
	Phase   *Phase
}

func SubscribeCurrent(c common.Context, s *subs.Subscription, email string) {
	s.Query = s.DB().Query().Where(kol.Equals{"UserId", []byte(email)})
	s.Call = func(i interface{}, op string) {
		members := i.([]*Member)
		states := []GameState{}
		phases := Phases{}
		for _, member := range members {
			if op == common.DeleteType {
				states = append(states, GameState{
					Game:    &Game{Id: member.GameId},
					Members: []MemberState{MemberState{Member: member}},
				})
			} else {
				game := &Game{Id: member.GameId}
				if err := s.DB().Get(game); err != nil {
					panic(err)
				}
				gameMembers := Members{}
				if err := s.DB().Query().Where(kol.Equals{"GameId", game.Id}).All(&gameMembers); err != nil {
					panic(err)
				}
				if !game.Ended {
					phases = nil
					if err := s.DB().Query().Where(kol.Equals{"GameId", game.Id}).All(&phases); err != nil {
						panic(err)
					}
					phase := &Phase{}
					if len(phases) > 0 {
						sort.Sort(phases)
						phase = &phases[0]
					} else {
						phase = nil
					}
					states = append(states, GameState{
						Game:    game,
						Members: gameMembers.toStates(c, game, email),
						Phase:   phase,
					})
				}
			}
		}
		s.Send(states, op)
	}
	s.Subscribe(new(Member))
}

func SubscribeOpen(c common.Context, s *subs.Subscription, email string) {
	s.Query = s.DB().Query().Where(kol.And{kol.Equals{"Closed", false}, kol.Equals{"Private", false}})
	s.Call = func(i interface{}, op string) {
		games := i.([]*Game)
		states := []GameState{}
		phases := Phases{}
		isMember := false
		for _, game := range games {
			members := Members{}
			if err := s.DB().Query().Where(kol.Equals{"GameId", game.Id}).All(&members); err != nil {
				panic(err)
			}
			isMember = false
			for _, m := range members {
				if string(m.UserId) == email {
					isMember = true
				}
			}
			if !isMember {
				phases = nil
				if err := s.DB().Query().Where(kol.Equals{"GameId", game.Id}).All(&phases); err != nil {
					panic(err)
				}
				phase := &Phase{}
				if len(phases) > 0 {
					sort.Sort(phases)
					phase = &phases[0]
				} else {
					phase = nil
				}
				states = append(states, GameState{
					Game:    game,
					Members: members.toStates(c, game, email),
					Phase:   phase,
				})
			}
		}
		s.Send(states, op)
	}
	s.Subscribe(new(Game))
}
