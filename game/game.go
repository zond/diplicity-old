package game

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/epoch"
	"github.com/zond/diplicity/user"
	"github.com/zond/godip/classical"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/state"
	"github.com/zond/kcwraps/kol"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Minutes int

type Games []*Game

func (self Games) SortAndLimit(f func(a, b *Game) bool, limit int) Games {
	sorted := SortedGames{
		Games:    self,
		LessFunc: f,
	}
	sort.Sort(sorted)
	if len(sorted.Games) > limit {
		return sorted.Games[:limit]
	}
	return sorted.Games
}

type SortedGames struct {
	Games    []*Game
	LessFunc func(a, b *Game) bool
}

func (self SortedGames) Len() int {
	return len(self.Games)
}

func (self SortedGames) Less(i, j int) bool {
	return self.LessFunc(self.Games[i], self.Games[j])
}

func (self SortedGames) Swap(i, j int) {
	self.Games[j], self.Games[i] = self.Games[i], self.Games[j]
}

type Game struct {
	Id kol.Id

	Closed             bool `kol:"index"`
	Private            bool `kol:"index"`
	State              common.GameState
	Variant            string
	AllocationMethod   string
	EndYear            int
	MinimumRanking     float64
	MaximumRanking     float64
	MinimumReliability float64

	SecretEmail    common.SecretFlag
	SecretNickname common.SecretFlag
	SecretNation   common.SecretFlag

	Deadlines map[dip.PhaseType]Minutes

	ChatFlags map[dip.PhaseType]common.ChatFlag

	NonCommitConsequences common.Consequence
	NMRConsequences       common.Consequence

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (self *Game) Disallows(u *user.User) bool {
	return u.Ranking < self.MinimumRanking || u.Ranking > self.MaximumRanking || u.Reliability() < self.MinimumReliability
}

func (self *Game) allocate(d *kol.DB, phase *Phase) (err error) {
	members, err := self.Members(d)
	if err != nil {
		return
	}
	switch self.AllocationMethod {
	case common.RandomString:
		for memberIndex, nationIndex := range rand.Perm(len(members)) {
			members[memberIndex].Nation = common.VariantMap[self.Variant].Nations[nationIndex]
		}
	case common.PreferencesString:
		prefs := make([][]dip.Nation, len(members))
		for index, member := range members {
			prefs[index] = member.PreferredNations
		}
		for index, nation := range optimizePreferences(prefs) {
			members[index].Nation = nation
		}
	default:
		return fmt.Errorf("Unknown allocation method %v", self.AllocationMethod)
	}
	for index, _ := range members {
		opts := dip.Options{}
		if opts, err = phase.Options(members[index].Nation); err != nil {
			return
		}
		members[index].Options = opts
		if len(opts) == 0 {
			members[index].Committed = true
			members[index].NoOrders = true
		} else {
			members[index].Committed = false
			members[index].NoOrders = false
		}
		if err = d.Set(&members[index]); err != nil {
			return
		}
	}
	return
}

func (self *Game) resolve(c common.SkinnyContext, phase *Phase) (err error) {
	// Check that we are in a phase where we CAN resolve
	if self.State != common.GameStateStarted {
		err = fmt.Errorf("%+v is not started", self)
		return
	}
	// Load our members
	members, err := self.Members(c.DB())
	if err != nil {
		return
	}
	// Load the godip state for the phase
	state, err := phase.State()
	if err != nil {
		return
	}
	// Load "now"
	epoch, err := epoch.Get(c.DB())
	if err != nil {
		return
	}
	// Just to limit runaway resolution to 100 phases.
	for i := 0; i < 100; i++ {
		// Resolve the phase!
		if err = state.Next(); err != nil {
			return
		}
		// Load the new godip phase from the state
		nextDipPhase := state.Phase()
		// Create a diplicity phase for the new phase
		nextPhase := &Phase{
			GameId:      self.Id,
			Ordinal:     phase.Ordinal + 1,
			Orders:      map[dip.Nation]map[dip.Province][]string{},
			Resolutions: map[dip.Province]string{},
			Season:      nextDipPhase.Season(),
			Year:        nextDipPhase.Year(),
			Type:        nextDipPhase.Type(),
			Deadline:    epoch + (time.Minute * time.Duration(self.Deadlines[nextDipPhase.Type()])),
		}
		// Set the new phase positions
		var resolutions map[dip.Province]error
		nextPhase.Units, nextPhase.SupplyCenters, nextPhase.Dislodgeds, nextPhase.Dislodgers, nextPhase.Bounces, resolutions = state.Dump()
		// Store the results of the previous godip phase in the previous diplicity phase
		for _, nationOrders := range phase.Orders {
			for prov, _ := range nationOrders {
				if res, found := resolutions[prov]; found && res != nil {
					phase.Resolutions[prov] = res.Error()
				} else {
					phase.Resolutions[prov] = "OK"
				}
			}
		}
		// Commit everyone that doesn't have any orders to give
		nrCommitted := 0
		for index, _ := range members {
			opts := dip.Options{}
			if opts, err = nextPhase.Options(members[index].Nation); err != nil {
				return
			}
			members[index].Options = opts
			if len(opts) == 0 {
				members[index].Committed = true
				members[index].NoOrders = true
				nrCommitted++
			} else {
				members[index].Committed = false
				members[index].NoOrders = false
			}
			if err = c.DB().Set(&members[index]); err != nil {
				return
			}
		}
		// Store the new phase
		if err = c.DB().Set(nextPhase); err != nil {
			return
		}
		// Mark the old phase as resolved, and save it
		phase.Resolved = true
		if err = c.DB().Set(phase); err != nil {
			return
		}
		// If everyone in the new phase isn't commited, schedule an auto resolve and break here.
		if nrCommitted < len(members) {
			if err = nextPhase.Schedule(c); err != nil {
				return
			}
			nextPhase.SendScheduleEmails(c, self)
			break
		}
		phase = nextPhase
	}
	return
}

func (self *Game) Describe(c common.SkinnyContext, trans common.Translator) (result string, err error) {
	switch self.State {
	case common.GameStateCreated:
		return trans.I(string(common.BeforeGamePhaseType))
	case common.GameStateStarted:
		var phase *Phase
		if _, phase, err = self.Phase(c.DB(), 0); err != nil {
			return
		}
		season := ""
		if season, err = trans.I(string(phase.Season)); err != nil {
			return
		}
		typ := ""
		if typ, err = trans.I(string(phase.Type)); err != nil {
			return
		}
		return trans.I("game_phase_description", season, phase.Year, typ)
	case common.GameStateEnded:
		return trans.I(string(common.AfterGamePhaseType))
	}
	err = fmt.Errorf("Unknown game state for %+v", self)
	return
}

func (self *Game) start(c common.SkinnyContext) (err error) {
	if self.State != common.GameStateCreated {
		err = fmt.Errorf("%+v is already started", self)
		return
	}
	self.State = common.GameStateStarted
	self.Closed = true
	if err = c.DB().Set(self); err != nil {
		return
	}
	var startState *state.State
	if self.Variant == common.ClassicalString {
		startState = classical.Start()
	} else {
		err = fmt.Errorf("Unknown variant %v", self.Variant)
		return
	}
	startPhase := startState.Phase()
	epoch, err := epoch.Get(c.DB())
	if err != nil {
		return
	}
	phase := &Phase{
		GameId:      self.Id,
		Ordinal:     0,
		Orders:      map[dip.Nation]map[dip.Province][]string{},
		Resolutions: map[dip.Province]string{},
		Season:      startPhase.Season(),
		Year:        startPhase.Year(),
		Type:        startPhase.Type(),
		Deadline:    epoch + (time.Minute * time.Duration(self.Deadlines[startPhase.Type()])),
	}
	phase.Units, phase.SupplyCenters, phase.Dislodgeds, phase.Dislodgers, phase.Bounces, _ = startState.Dump()
	if err = c.DB().Set(phase); err != nil {
		return
	}
	if err = self.allocate(c.DB(), phase); err != nil {
		return
	}
	if err = phase.Schedule(c); err != nil {
		return
	}
	phase.SendScheduleEmails(c, self)
	return
}

func (self *Game) Updated(d *kol.DB, old *Game) {
	if old != self {
		members, err := self.Members(d)
		if err == nil {
			for _, member := range members {
				d.EmitUpdate(&member)
			}
		}
	}
}

func (self *Game) Phases(d *kol.DB) (result Phases, err error) {
	err = d.Query().Where(kol.Equals{"GameId", self.Id}).All(&result)
	return
}

func (self *Game) Phase(d *kol.DB, ordinal int) (result, last *Phase, err error) {
	phases, err := self.Phases(d)
	if err != nil {
		return
	}
	for index, _ := range phases {
		if last == nil || phases[index].Ordinal > last.Ordinal {
			last = &phases[index]
		}
		if phases[index].Ordinal == ordinal {
			result = &phases[index]
		}
	}
	return
}

func (self *Game) Members(d *kol.DB) (result Members, err error) {
	if err = d.Query().Where(kol.Equals{"GameId", self.Id}).All(&result); err != nil {
		return
	}
	sort.Sort(result)
	return
}

func (self *Game) UnseenMessages(d *kol.DB, viewer kol.Id) (result map[string]int, err error) {
	msgs, err := self.Messages(d)
	if err != nil {
		return
	}
	result = map[string]int{}
	for _, msg := range msgs {
		if msg.RecipientIds[viewer.String()] && !msg.SeenBy[viewer.String()] {
			result[msg.ChannelId()]++
		}
	}
	return
}

func (self *Game) ToState(d *kol.DB, members Members, member *Member) (result GameState, err error) {
	_, phase, err := self.Phase(d, 0)
	if err != nil {
		return
	}
	ordinal := 0
	if phase != nil {
		ordinal = phase.Ordinal
	}
	return self.toStateWithPhase(d, members, member, phase.redact(member), ordinal)
}

func (self *Game) ToStateWithPhaseOrdinal(d *kol.DB, members Members, member *Member, ordinal int) (result GameState, err error) {
	phase, last, err := self.Phase(d, ordinal)
	if err != nil {
		return
	}
	if phase == nil {
		err = fmt.Errorf("No phase with ordinal %v in %v", ordinal, self)
		return
	}
	if last == phase {
		phase = phase.redact(member)
	}
	return self.toStateWithPhase(d, members, member, phase, last.Ordinal)
}

func (self *Game) toStateWithPhase(d *kol.DB, members Members, member *Member, phase *Phase, phases int) (result GameState, err error) {
	email := ""
	if member != nil {
		email = string(member.UserId)
	}
	memberStates, err := members.ToStates(d, self, email, false)
	if err != nil {
		return
	}
	unseen := map[string]int{}
	if member != nil {
		unseen, err = self.UnseenMessages(d, member.Id)
		if err != nil {
			return
		}
	}
	var timeLeft time.Duration
	if phase != nil {
		timeLeft, err = epoch.Get(d)
		if err != nil {
			return
		}
		timeLeft = phase.Deadline - timeLeft
	}
	result = GameState{
		Game:           self,
		UnseenMessages: unseen,
		Members:        memberStates,
		TimeLeft:       timeLeft,
		Phase:          phase,
		Phases:         phases,
	}
	return
}

func (self *Game) Messages(d *kol.DB) (result Messages, err error) {
	if err = d.Query().Where(kol.Equals{"GameId", self.Id}).All(&result); err != nil {
		return
	}
	sort.Sort(result)
	return
}

func (self *Game) Member(d *kol.DB, email string) (result *Member, err error) {
	var member Member
	var found bool
	if found, err = d.Query().Where(kol.And{kol.Equals{"GameId", self.Id}, kol.Equals{"UserId", kol.Id(email)}}).First(&member); found && err == nil {
		result = &member
	}
	return
}

func (self *Game) Users(d *kol.DB) (result user.Users, err error) {
	members, err := self.Members(d)
	if err != nil {
		return
	}
	result = make(user.Users, len(members))
	for index, member := range members {
		user := user.User{Id: member.UserId}
		if err = d.Get(&user); err != nil {
			return
		}
		result[index] = user
	}
	return
}
