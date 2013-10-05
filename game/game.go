package game

import (
	"encoding/base64"
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	"github.com/zond/godip/classical"
	"github.com/zond/godip/classical/orders"
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
	members := self.Members(d)
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
	if self.Started {
		return fmt.Errorf("%+v is already started", self)
	}
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
		GameId:  self.Id,
		Ordinal: 0,
		Season:  startPhase.Season(),
		Year:    startPhase.Year(),
		Type:    startPhase.Type(),
	}
	phase.Units, phase.SupplyCenters, phase.Dislodgeds, phase.Dislodgers, phase.Bounces = startState.Dump()
	return d.Set(phase)
}

func DeleteMember(c common.Context, gameId, email string) error {
	return c.DB().Transact(func(d *kol.DB) error {
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
		if game.Started {
			return fmt.Errorf("%+v already started", game)
		}
		member := Member{}
		if _, err := d.Query().Where(kol.And{kol.Equals{"GameId", base64DecodedId}, kol.Equals{"UserId", []byte(email)}}).First(&member); err != nil {
			return err
		}
		if err := d.Del(&member); err != nil {
			return err
		}
		left := game.Members(d)
		if len(left) == 0 {
			if err := d.Del(&game); err != nil {
				return err
			}
		}
		return nil
	})
}

func AddMember(c common.Context, j subs.JSON, email string) error {
	var state GameState
	j.Overwrite(&state)
	return c.DB().Transact(func(d *kol.DB) error {
		game := Game{Id: state.Game.Id}
		if err := d.Get(&game); err != nil {
			return fmt.Errorf("Game not found")
		}
		if game.Started {
			return fmt.Errorf("%+v already started")
		}
		variant, found := common.VariantMap[game.Variant]
		if !found {
			return fmt.Errorf("Unknown variant %v", game.Variant)
		}
		already := game.Members(d)
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
	})
}

func Create(c common.Context, j subs.JSON, creator string) error {
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

	if _, found := common.VariantMap[game.Variant]; !found {
		return fmt.Errorf("Unknown variant for %+v", game)
	}

	if _, found := common.AllocationMethodMap[game.AllocationMethod]; !found {
		return fmt.Errorf("Unknown allocation method for %+v", game)
	}

	member := &Member{
		UserId:           []byte(creator),
		PreferredNations: state.Members[0].PreferredNations,
	}
	return c.DB().Transact(func(d *kol.DB) error {
		if err := d.Set(game); err != nil {
			return err
		}
		member.GameId = game.Id
		return d.Set(member)
	})
}

func (self *Game) Updated(d *kol.DB, old *Game) {
	for _, member := range self.Members(d) {
		d.EmitUpdate(&member)
	}
}

func (self *Game) LastPhase(d *kol.DB) (result *Phase) {
	var phases Phases
	d.Query().Where(kol.Equals{"GameId", self.Id}).All(&phases)
	if len(phases) > 0 {
		sort.Sort(phases)
		result = &phases[0]
	}
	return
}

func (self *Game) Members(d *kol.DB) (result Members) {
	d.Query().Where(kol.Equals{"GameId", self.Id}).All(&result)
	return
}

func (self *Game) Member(d *kol.DB, email string) (result *Member, err error) {
	var member Member
	if _, err = d.Query().Where(kol.And{kol.Equals{"GameId", self.Id}, kol.Equals{"UserId", []byte(email)}}).First(&member); err == nil {
		result = &member
	}
	return
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
	Dislodgeds    map[dip.Province]dip.Unit
	Dislodgers    map[dip.Province]dip.Province
	Bounces       map[dip.Province]map[dip.Province]bool
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
		if string(cpy.UserId) != email && g.SecretNation {
			cpy.Nation = ""
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
		return s.Send(states, op)
	}
	return s.Subscribe(new(Member))
}

func SubscribeGame(c common.Context, s *subs.Subscription, gameId, email string) error {
	urlDecodedId, err := url.QueryUnescape(gameId)
	if err != nil {
		return err
	}
	base64DecodedId, err := base64.StdEncoding.DecodeString(urlDecodedId)
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
		for _, game := range games {
			members := game.Members(c.DB())
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
		return s.Send(states, op)
	}
	return s.Subscribe(new(Game))
}

func GetValidOrders(c common.Context, gameId, province, email string) (result dip.Options, err error) {
	var base64DecodedId []byte
	base64DecodedId, err = base64.StdEncoding.DecodeString(gameId)
	if err != nil {
		return
	}
	game := Game{Id: base64DecodedId}
	if err = c.DB().Get(&game); err != nil {
		return
	}
	var member *Member
	member, err = game.Member(c.DB(), email)
	if err != nil {
		return
	}
	phase := game.LastPhase(c.DB())
	if phase == nil {
		err = fmt.Errorf("No phase for %+v found", game)
		return
	}
	state := classical.Blank(classical.Phase(phase.Year, phase.Season, phase.Type))
	state.Load(
		phase.Units,
		phase.SupplyCenters,
		phase.Dislodgeds,
		phase.Dislodgers,
		phase.Bounces,
	)
	nation, options, found := state.Options(orders.Types(), dip.Province(province))
	if found && nation == member.Nation {
		result = options
	}
	return
}
