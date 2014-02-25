package game

import (
	"time"

	"github.com/zond/godip/classical"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/state"
	"github.com/zond/kcwraps/kol"
)

type Phase struct {
	Id     kol.Id
	GameId kol.Id `kol:"index"`

	Season   dip.Season
	Year     int
	Type     dip.PhaseType
	Ordinal  int
	Resolved bool

	Units         map[dip.Province]dip.Unit
	Orders        map[dip.Nation]map[dip.Province][]string
	SupplyCenters map[dip.Province]dip.Nation
	Dislodgeds    map[dip.Province]dip.Unit
	Dislodgers    map[dip.Province]dip.Province
	Bounces       map[dip.Province]map[dip.Province]bool

	Committed map[dip.Nation]bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (self *Phase) Updated(d *kol.DB, old *Phase) {
	g := Game{Id: self.GameId}
	if err := d.Get(&g); err != nil {
		panic(err)
	}
	d.EmitUpdate(&g)
}

func (self Phase) redact(member *Member) *Phase {
	for nat, _ := range self.Committed {
		if member == nil || member.Nation != nat {
			delete(self.Committed, nat)
		}
	}
	if !self.Resolved {
		for nat, _ := range self.Orders {
			if member == nil || member.Nation != nat {
				delete(self.Orders, nat)
			}
		}
	}
	return &self
}

func (self *Phase) GetState() *state.State {
	return classical.Blank(classical.Phase(
		self.Year,
		self.Season,
		self.Type,
	)).Load(
		self.Units,
		self.SupplyCenters,
		self.Dislodgeds,
		self.Dislodgers,
		self.Bounces,
	)
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
