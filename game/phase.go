package game

import (
	"fmt"
	"time"

	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/epoch"
	"github.com/zond/diplicity/user"
	"github.com/zond/godip/classical"
	"github.com/zond/godip/classical/orders"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/state"
	"github.com/zond/kcwraps/kol"
)

func ScheduleUnresolvedPhases(c common.SkinnyContext) (err error) {
	unresolved := Phases{}
	if err = c.DB().Query().Where(kol.Equals{"Resolved", false}).All(&unresolved); err != nil {
		return
	}
	for _, phase := range unresolved {
		g := &Game{Id: phase.GameId}
		if err = c.DB().Get(g); err != nil {
			return
		}
		members := Members{}
		if members, err = g.Members(c.DB()); err != nil {
			return
		}
		for index, _ := range members {
			member := &members[index]
			opts := dip.Options{}
			if opts, err = phase.Options(member.Nation); err != nil {
				return
			}
			member.Options = opts
			if err = c.DB().Set(member); err != nil {
				return
			}
			c.Infof("### created options for %v", member.Nation)
		}
		phase.Schedule(c)
	}
	return
}

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

	Deadline time.Duration

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (self *Phase) ShortString() string {
	return fmt.Sprintf("%v %v, %v", self.Season, self.Year, self.Type)
}

func (self *Phase) Schedule(c common.SkinnyContext) (err error) {
	if !self.Resolved {
		var ep time.Duration
		ep, err = epoch.Get(c.DB())
		if err != nil {
			return
		}
		timeout := self.Deadline - ep
		c.BetweenTransactions(func(c common.SkinnyContext) {
			time.AfterFunc(timeout, func() {
				c.Infof("Auto resolving %v due to timeout", self.GameId)
				if err := c.Transact(func(c common.SkinnyContext) (err error) {
					if err = c.DB().Get(self); err != nil {
						return
					}
					if self.Resolved {
						c.Infof("%v was already resolved, skipping", self.Id)
						return
					}
					game := &Game{Id: self.GameId}
					if err = c.DB().Get(game); err != nil {
						return
					}
					return game.resolve(c, self)
				}); err != nil {
					c.Errorf("Unable to resolve %+v: %v", self, err)
				}
			})
			c.Infof("Scheduled resolution of %v in %v", self.GameId, timeout)
		})
	}
	return
}

func (self *Phase) SendScheduleEmails(c common.SkinnyContext, game *Game) {
	from := fmt.Sprintf("diplicity <%v>", c.MailAddress())

	members, err := game.Members(c.DB())
	for _, member := range members {
		user := &user.User{Id: member.UserId}
		if err = c.DB().Get(user); err != nil {
			return
		}
		to := fmt.Sprintf("%v <%v>", member.Nation, user.Email)
		if !user.PhaseEmailDisabled && !c.IsSubscribing(user.Email, fmt.Sprintf("/games/%v", game.Id)) {
			unsubTag := &common.UnsubscribeTag{
				T: common.UnsubscribePhaseEmail,
				U: user.Id,
			}
			unsubTag.H = unsubTag.Hash(c.Secret())
			encodedUnsubTag, err := unsubTag.Encode()
			if err != nil {
				c.Errorf("Failed to encode %+v: %v", unsubTag, err)
				return
			}
			contextLink, err := user.I("To see this in context: http://%v/games/%v", user.DiplicityHost, self.GameId)
			if err != nil {
				c.Errorf("Failed translating context link: %v", err)
				return
			}
			unsubLink, err := user.I("To unsubscribe: http://%v/unsubscribe/%v", user.DiplicityHost, encodedUnsubTag)
			if err != nil {
				c.Errorf("Failed translating unsubscribe link: %v", err)
				return
			}
			text, err := user.I("A new phase has been created")
			if err != nil {
				c.Errorf("Failed translating: %v", err)
				return
			}
			subject, err := game.Describe(c, user)
			if err != nil {
				c.Errorf("Failed describing: %v", err)
				return
			}
			body := fmt.Sprintf(common.EmailTemplate, text, contextLink, unsubLink)
			if c.Env() == "development" {
				c.Infof("Would have sent\nFrom: %#v\nTo: %#v\nSubject: %#v\n%v", from, to, subject, body)
			} else {
				if err := c.SendMail(from, subject, body, to); err == nil {
					c.Infof("Sent\nFrom: %#v\nTo: %#v\nSubject: %#v\n%v", from, to, subject, body)
				} else {
					c.Errorf("Unable to send %#v/%#v from %#v to %#v: %v", subject, body, from, to, err)
				}
			}
		}
	}
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
	if !self.Resolved {
		for nat, _ := range self.Orders {
			if member == nil || member.Nation != nat {
				delete(result.Orders, nat)
			}
		}
	}
	return &result
}

func (self *Phase) Options(nation dip.Nation) (result dip.Options, err error) {
	state, err := self.State()
	if err != nil {
		return
	}
	result = state.Phase().Options(state, nation)
	return
}

func (self *Phase) State() (result *state.State, err error) {
	parsedOrders, err := orders.ParseAll(self.Orders)
	if err != nil {
		return
	}
	units := map[dip.Province]dip.Unit{}
	for prov, unit := range self.Units {
		units[prov] = unit
	}
	orders := map[dip.Nation]map[dip.Province][]string{}
	for nat, ord := range self.Orders {
		orders[nat] = ord
	}
	supplyCenters := map[dip.Province]dip.Nation{}
	for prov, nat := range self.SupplyCenters {
		supplyCenters[prov] = nat
	}
	dislodgeds := map[dip.Province]dip.Unit{}
	for prov, unit := range self.Dislodgeds {
		dislodgeds[prov] = unit
	}
	dislodgers := map[dip.Province]dip.Province{}
	for k, v := range self.Dislodgers {
		dislodgers[k] = v
	}
	bounces := map[dip.Province]map[dip.Province]bool{}
	for prov, b := range self.Bounces {
		bounces[prov] = b
	}
	result = classical.Blank(classical.Phase(
		self.Year,
		self.Season,
		self.Type,
	)).Load(
		units,
		supplyCenters,
		dislodgeds,
		dislodgers,
		bounces,
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
