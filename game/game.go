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

type Game struct {
	Id kol.Id

	Closed             bool `kol:"index"`
	State              common.GameState
	Variant            string
	AllocationMethod   string
	SecretEmail        common.SecretFlag
	SecretNickname     common.SecretFlag
	SecretNation       common.SecretFlag
	EndYear            int
	MinimumRanking     float64
	MaximumRanking     float64
	MinimumReliability float64
	Private            bool `kol:"index"`

	Deadlines map[dip.PhaseType]Minutes

	ChatFlags map[dip.PhaseType]common.ChatFlag

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (self *Game) Disallows(u *user.User) bool {
	return u.Ranking < self.MinimumRanking || u.Ranking > self.MaximumRanking || u.Reliability() < self.MinimumReliability
}

func (self *Game) allocate(d *kol.DB) error {
	members, err := self.Members(d)
	if err != nil {
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

func (self *Game) resolve(c common.SkinnyContext, phase *Phase) (err error) {
	// Check that we are in a phase where we CAN resolve
	if self.State != common.GameStateStarted {
		err = fmt.Errorf("%+v is not started", self)
		return
	}
	// Check that we have a valid variant
	variant, found := common.VariantMap[self.Variant]
	if !found {
		err = fmt.Errorf("Unrecognized variant for %+v", self)
		return
	}
	var possibleSources []dip.Province
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
			Committed:   map[dip.Nation]bool{},
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
		for _, nation := range variant.Nations {
			possibleSources, err = nextPhase.PossibleSources(nation)
			if err != nil {
				return
			}
			if len(possibleSources) == 0 {
				nextPhase.Committed[nation] = true
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
		if len(nextPhase.Committed) < len(variant.Nations) {
			if err = nextPhase.Schedule(c); err != nil {
				return
			}
			break
		}
		phase = nextPhase
	}
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
	if err = self.allocate(c.DB()); err != nil {
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
		Committed:   map[dip.Nation]bool{},
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
	if err = phase.Schedule(c); err != nil {
		return
	}
	return
}

func (self *Game) Updated(d *kol.DB, old *Game) {
	members, err := self.Members(d)
	if err == nil {
		for _, member := range members {
			d.EmitUpdate(&member)
		}
	}
}

func (self *Game) Phases(d *kol.DB) (result Phases, err error) {
	err = d.Query().Where(kol.Equals{"GameId", self.Id}).All(&result)
	return
}

func (self *Game) LastPhase(d *kol.DB) (result *Phase, err error) {
	phases, err := self.Phases(d)
	if err != nil {
		return
	}
	if len(phases) > 0 {
		sort.Sort(phases)
		result = &phases[0]
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
