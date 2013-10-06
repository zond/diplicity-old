package game

import (
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/godip/classical"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/state"
	"github.com/zond/kcwraps/kol"
	"math/rand"
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
