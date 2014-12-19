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

const (
	RankingBlind = 1.0 / 16.0
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
	Games    Games
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

	Closed             bool             `kol:"index"`
	Private            bool             `kol:"index"`
	State              common.GameState `kol:"index"`
	EndReason          common.EndReason
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

	Ranking bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (self *Game) Disallows(u *user.User) bool {
	return (self.MinimumRanking != 0 && u.Ranking < self.MinimumRanking) ||
		(self.MaximumRanking != 0 && u.Ranking > self.MaximumRanking) ||
		u.Reliability() < self.MinimumReliability
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

func (self *Game) endPhaseConsequences(c common.SkinnyContext, phase *Phase, member *Member, opts dip.Options, waitFor, active, nonSurrendering *[]*Member) (err error) {
	surrender := false
	if !member.Committed {
		alreadyHitReliability := false
		if (self.NonCommitConsequences & common.ReliabilityHit) == common.ReliabilityHit {
			if err = member.ReliabilityDelta(c.DB(), -1); err != nil {
				return
			}
			c.Infof("Increased MISSED deadlines for %#v by one because %+v, %+v and %+v", string(member.UserId), self, member, phase)
			alreadyHitReliability = true
		}
		if (self.NonCommitConsequences & common.NoWait) == common.NoWait {
			c.Infof("Setting %#v to NoWait because of %+v, %+v and %+v", string(member.UserId), self, member, phase)
			member.NoWait = true
		}
		if (self.NonCommitConsequences & common.Surrender) == common.Surrender {
			c.Infof("Setting %#v to Surrender because of %+v, %+v and %+v", string(member.UserId), self, member, phase)
			surrender = true
		}
		if len(phase.Orders[member.Nation]) == 0 {
			if !alreadyHitReliability && (self.NMRConsequences&common.ReliabilityHit) == common.ReliabilityHit {
				if err = member.ReliabilityDelta(c.DB(), -1); err != nil {
					return
				}
				c.Infof("Increased MISSED deadlines for %#v by one because %+v, %+v and %+v", string(member.UserId), self, member, phase)
			}
			if (self.NMRConsequences & common.NoWait) == common.NoWait {
				c.Infof("Setting %#v to NoWait because of %+v, %+v and %+v", string(member.UserId), self, member, phase)
				member.NoWait = true
			}
			if (self.NMRConsequences & common.Surrender) == common.Surrender {
				c.Infof("Setting %#v to Surrender because of %+v, %+v and %+v", string(member.UserId), self, member, phase)
				surrender = true
			}
		}
	} else {
		if (self.NonCommitConsequences&common.ReliabilityHit) == common.ReliabilityHit || (self.NMRConsequences&common.ReliabilityHit) == common.ReliabilityHit {
			if err = member.ReliabilityDelta(c.DB(), 1); err != nil {
				return
			}
			c.Infof("Increased HELD deadlines for %#v by one because %+v, %+v and %+v", string(member.UserId), self, member, phase)
		}
	}
	if !surrender {
		*nonSurrendering = append(*nonSurrendering, member)
	}
	member.Options = opts
	if member.NoWait {
		member.Committed = false
		member.NoOrders = false
	} else {
		*active = append(*active, member)
		if len(opts) == 0 {
			member.Committed = true
			member.NoOrders = true
		} else {
			*waitFor = append(*waitFor, member)
			member.Committed = false
			member.NoOrders = false
		}
	}
	if err = c.DB().Set(member); err != nil {
		return
	}
	return
}

func (self *Game) end(c common.SkinnyContext, phase *Phase, members Members, winner *Member, reason common.EndReason) (err error) {
	self.EndReason = reason
	self.State = common.GameStateEnded
	if err = c.DB().Set(self); err != nil {
		return
	}
	phase.Resolved = true
	if err = c.DB().Set(phase); err != nil {
		return
	}
	if self.Ranking && winner != nil {
		pot := 0.0
		spend := 0.0
		for index, _ := range members {
			if !members[index].Id.Equals(winner.Id) {
				user := &user.User{Id: members[index].UserId}
				if err = c.DB().Get(user); err != nil {
					return
				}
				spend = user.Ranking * RankingBlind
				pot += spend
				user.Ranking -= spend
				if err = c.DB().Set(user); err != nil {
					return
				}
			}
		}
		winnerUser := &user.User{Id: winner.UserId}
		if err = c.DB().Get(winnerUser); err != nil {
			return
		}
		winnerUser.Ranking += pot
		if err = c.DB().Set(winnerUser); err != nil {
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
		waitFor := []*Member{}
		active := []*Member{}
		nonSurrendering := []*Member{}
		for index, _ := range members {
			opts := dip.Options{}
			if opts, err = nextPhase.Options(members[index].Nation); err != nil {
				return
			}
			if err = self.endPhaseConsequences(c, phase, &members[index], opts, &waitFor, &active, &nonSurrendering); err != nil {
				return
			}
		}

		// Mark the old phase as resolved, and save it
		phase.Resolved = true
		if err = c.DB().Set(phase); err != nil {
			return
		}

		// If we have a solo victor, end and return
		if winner := nextDipPhase.Winner(state); winner != nil {
			var winnerMember *Member
			for _, member := range members {
				if member.Nation == *winner {
					winnerMember = &member
					break
				}
			}
			if winnerMember == nil {
				err = fmt.Errorf("None of %+v has nation %#v??", members, *winner)
				return
			}
			if err = self.end(c, nextPhase, members, winnerMember, common.SoloVictory(*winner)); err != nil {
				return
			}
			return
		}

		// End the game now if nobody is active anymore
		if len(active) == 0 {
			if err = self.end(c, nextPhase, members, nil, common.ZeroActiveMembers); err != nil {
				return
			}
			return
		}

		// End the game now if only one player isn't surrendering
		if len(nonSurrendering) == 1 {
			if err = self.end(c, nextPhase, members, nonSurrendering[0], common.SoloVictory(nonSurrendering[0].Nation)); err != nil {
				return
			}
			return
		}

		// Store the next phase
		if err = c.DB().Set(nextPhase); err != nil {
			return
		}

		// If there is anyone we need to wait for, schedule an auto resolve and return here.
		if len(waitFor) > 0 {
			if err = nextPhase.Schedule(c); err != nil {
				return
			}
			nextPhase.SendStartedEmails(c, self)
			return
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
		if startState, err = classical.Start(); err != nil {
			return
		}
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
	phase.SendStartedEmails(c, self)
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
