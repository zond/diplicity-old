package game

import (
	"time"

	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/epoch"
	"github.com/zond/godip/classical"
	"github.com/zond/godip/classical/orders"
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
	Resolved bool `kol:"index"`

	Units         map[dip.Province]dip.Unit
	Orders        map[dip.Nation]map[dip.Province][]string
	SupplyCenters map[dip.Province]dip.Nation
	Dislodgeds    map[dip.Province]dip.Unit
	Dislodgers    map[dip.Province]dip.Province
	Bounces       map[dip.Province]map[dip.Province]bool
	Resolutions   map[dip.Province]string

	Committed map[dip.Nation]bool

	Deadline time.Duration

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (self *Phase) Schedule(c common.SkinnyContext) (err error) {
	if !self.Resolved {
		var ep time.Duration
		ep, err = epoch.Get(c.DB())
		if err != nil {
			return
		}
		timeout := self.Deadline - ep
		time.AfterFunc(timeout, func() {
			if err := c.Transact(func(c common.SkinnyContext) (err error) {
				if err = c.DB().Get(self); err != nil {
					return
				}
				if !self.Resolved {
					ep, err = epoch.Get(c.DB())
					if err != nil {
						return
					}
					if ep > self.Deadline {
						game := &Game{Id: self.GameId}
						if err = c.DB().Get(game); err != nil {
							return
						}
						return game.resolve(c, self)
					}
				}
				return
			}); err != nil {
				c.Errorf("Unable to resolve %+v: %v", self, err)
			}
		})
		c.Infof("Scheduled resolution of %v in %v", self.GameId, timeout)
	}
	return
}

func (self *Phase) Game(d *kol.DB) (result *Game, err error) {
	result = &Game{Id: self.GameId}
	err = d.Get(result)
	return
}

func (self *Phase) Updated(d *kol.DB, old *Phase) {
	g := Game{Id: self.GameId}
	if err := d.Get(&g); err != nil {
		panic(err)
	}
	d.EmitUpdate(&g)
}

func (self *Phase) redact(member *Member) *Phase {
	if self == nil {
		return nil
	}
	result := *self
	for nat, _ := range self.Committed {
		if member == nil || member.Nation != nat {
			delete(result.Committed, nat)
		}
	}
	if !self.Resolved {
		for nat, _ := range self.Orders {
			if member == nil || member.Nation != nat {
				delete(result.Orders, nat)
			}
		}
	}
	return &result
}

func (self *Phase) PossibleSources(nation dip.Nation) (result []dip.Province, err error) {
	state, err := self.State()
	if err != nil {
		return
	}
	result = state.Phase().PossibleSources(state, nation)
	return
}

func (self *Phase) State() (result *state.State, err error) {
	parsedOrders, err := orders.ParseAll(self.Orders)
	if err != nil {
		return
	}
	result = classical.Blank(classical.Phase(
		self.Year,
		self.Season,
		self.Type,
	)).Load(
		self.Units,
		self.SupplyCenters,
		self.Dislodgeds,
		self.Dislodgers,
		self.Bounces,
		parsedOrders,
	)
	return
}

type Phases []Phase

func (self Phases) Len() int {
	return len(self)
}

func (self Phases) Less(j, i int) bool {
	return self[i].Ordinal < self[j].Ordinal
}

func (self Phases) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
