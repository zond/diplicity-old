package state

import (
	"fmt"
	"github.com/zond/godip/common"
)

func New(graph common.Graph, phase common.Phase, backupRule common.BackupRule) *State {
	return &State{
		graph:         graph,
		phase:         phase,
		backupRule:    backupRule,
		orders:        make(map[common.Province]common.Adjudicator),
		units:         make(map[common.Province]common.Unit),
		dislodgeds:    make(map[common.Province]common.Unit),
		supplyCenters: make(map[common.Province]common.Nation),
		dislodgers:    make(map[common.Province]common.Province),
		bounces:       make(map[common.Province]map[common.Province]bool),
	}
}

type movement struct {
	src            common.Province
	dst            common.Province
	unit           common.Unit
	preventRetreat bool
}

func (self *movement) prepare(s *State) {
	var ok bool
	if self.unit, self.src, ok = s.Unit(self.src); !ok {
		panic(fmt.Errorf("No unit at %v?", self.src))
	} else {
		s.RemoveUnit(self.src)
	}
	common.Logf("Lifted %v from %v", self.unit, self.src)
}

func (self *movement) execute(s *State) {
	if dislodged, prov, ok := s.Unit(self.dst); ok {
		s.RemoveUnit(prov)
		s.SetDislodged(prov, dislodged)
		if self.preventRetreat {
			s.SetDislodger(self.src, prov)
		}
		common.Logf("Dislodged %v from %v", dislodged, self.dst)
	}
	s.SetUnit(self.dst, self.unit)
	common.Logf("Dropped %v in %v", self.unit, self.dst)
}

type State struct {
	orders        map[common.Province]common.Adjudicator
	units         map[common.Province]common.Unit
	dislodgeds    map[common.Province]common.Unit
	supplyCenters map[common.Province]common.Nation
	graph         common.Graph
	phase         common.Phase
	backupRule    common.BackupRule
	resolutions   map[common.Province]error
	dislodgers    map[common.Province]common.Province
	movements     []*movement
	bounces       map[common.Province]map[common.Province]bool
}

func (self *State) resolver() *resolver {
	return &resolver{
		State:     self,
		guesses:   make(map[common.Province]error),
		resolving: make(map[common.Province]bool),
	}
}

func (self *State) Graph() common.Graph {
	return self.graph
}

func (self *State) Find(filter common.StateFilter) (provinces []common.Province, orders []common.Order, units []*common.Unit) {
	visitedProvinces := make(map[common.Province]bool)
	for prov, unit := range self.units {
		visitedProvinces[prov] = true
		var order common.Order
		var ok bool
		if order, _, ok = self.Order(prov); !ok {
			order = nil
		}
		unitCopy := unit
		if filter(prov, order, &unit) {
			provinces = append(provinces, prov)
			orders = append(orders, order)
			units = append(units, &unitCopy)
		}
	}
	for prov, order := range self.orders {
		if !visitedProvinces[prov] {
			if filter(prov, order, nil) {
				provinces = append(provinces, prov)
				orders = append(orders, order)
				units = append(units, nil)
			}
		}
	}
	return
}

func (self *State) Next() (err error) {
	/*
	   Sanitize orders.
	*/
	self.resolutions = make(map[common.Province]error)
	for prov, order := range self.orders {
		if err := order.Validate(self); err != nil {
			self.resolutions[prov] = err
			delete(self.orders, prov)
			common.Logf("Deleted %v due to %v", prov, err)
		}
	}

	/*
		Add hold to units missing orders.
	*/
	for prov, _ := range self.units {
		if _, ok := self.orders[prov]; !ok {
			if def := self.phase.DefaultOrder(prov); def != nil {
				self.orders[prov] = def
			}
		}
	}

	/*
	   Adjudicate orders.
	*/
	for prov, _ := range self.orders {
		err := self.resolver().Resolve(prov)
		self.resolutions[prov] = err
	}

	/*
	   Execute orders.
	*/
	self.movements = nil
	for prov, order := range self.orders {
		if err, ok := self.resolutions[prov]; ok && err == nil {
			order.Execute(self.resolver())
		}
	}
	self.orders = make(map[common.Province]common.Adjudicator)

	/*
	   Execute movements.
	*/
	for _, movement := range self.movements {
		movement.prepare(self)
	}
	for _, movement := range self.movements {
		movement.execute(self)
	}

	/*
	   Change phase.
	*/
	self.phase.PostProcess(self.resolver())
	self.phase = self.phase.Next()
	return
}

func (self *State) Phase() common.Phase {
	return self.phase
}

// Bulk setters

func (self *State) SetOrders(orders map[common.Province]common.Adjudicator) *State {
	self.orders = make(map[common.Province]common.Adjudicator)
	for prov, order := range orders {
		self.SetOrder(prov, order)
	}
	return self
}

func (self *State) SetUnits(units map[common.Province]common.Unit) *State {
	self.units = make(map[common.Province]common.Unit)
	for prov, unit := range units {
		self.SetUnit(prov, unit)
	}
	return self
}

func (self *State) SetDislodgeds(dislodgeds map[common.Province]common.Unit) *State {
	self.dislodgeds = make(map[common.Province]common.Unit)
	for prov, unit := range dislodgeds {
		self.SetDislodged(prov, unit)
	}
	return self
}

func (self *State) SetSupplyCenters(supplyCenters map[common.Province]common.Nation) *State {
	self.supplyCenters = supplyCenters
	return self
}

func (self *State) ClearBounces() {
	self.bounces = make(map[common.Province]map[common.Province]bool)
}

func (self *State) ClearDislodgers() {
	self.dislodgers = make(map[common.Province]common.Province)
}

// Singular setters

func (self *State) SetDislodger(attacker, victim common.Province) {
	self.dislodgers[attacker.Super()] = victim.Super()
}

func (self *State) AddBounce(src, dst common.Province) {
	if existing, ok := self.bounces[dst.Super()]; ok {
		existing[src.Super()] = true
	} else {
		self.bounces[dst.Super()] = map[common.Province]bool{
			src.Super(): true,
		}
	}
}

func (self *State) SetResolution(p common.Province, err error) {
	self.resolutions[p] = err
}

func (self *State) SetSC(p common.Province, n common.Nation) {
	self.supplyCenters[p] = n
}

func (self *State) SetDislodged(prov common.Province, unit common.Unit) {
	if found, _, ok := self.Dislodged(prov); ok {
		panic(fmt.Errorf("%v is already at %v", found, prov))
	}
	self.dislodgeds[prov] = unit
}

func (self *State) SetUnit(prov common.Province, unit common.Unit) {
	if found, _, ok := self.Unit(prov); ok {
		panic(fmt.Errorf("%v is already at %v", found, prov))
	}
	self.units[prov] = unit
}

func (self *State) SetOrder(prov common.Province, order common.Adjudicator) {
	if found, _, ok := self.Order(prov); ok {
		panic(fmt.Errorf("%v is already at %v", found, prov))
	}
	self.orders[prov] = order
}

func (self *State) RemoveUnit(prov common.Province) {
	if _, p, ok := self.Unit(prov); ok {
		delete(self.units, p)
	}
}

func (self *State) RemoveDislodged(prov common.Province) {
	if _, p, ok := self.Dislodged(prov); ok {
		delete(self.dislodgeds, p)
	}
}

// Bulk getters

func (self *State) Resolutions() map[common.Province]error {
	return self.resolutions
}

func (self *State) SupplyCenters() map[common.Province]common.Nation {
	return self.supplyCenters
}

func (self *State) Units() map[common.Province]common.Unit {
	return self.units
}

func (self *State) Dislodgeds() map[common.Province]common.Unit {
	return self.dislodgeds
}

func (self *State) Orders() map[common.Province]common.Adjudicator {
	return self.orders
}

// Singular getters, will search all coasts of a province

func (self *State) Bounce(src, dst common.Province) bool {
	if sources, ok := self.bounces[dst.Super()]; ok {
		if dislodger, ok := self.dislodgers[dst.Super()]; ok {
			if len(sources) == 1 && sources[dislodger.Super()] {
				return false
			}
		}
		return true
	}
	if self.dislodgers[dst.Super()] == src.Super() {
		return true
	}
	return false
}

func (self *State) Dislodged(prov common.Province) (u common.Unit, p common.Province, ok bool) {
	if u, ok = self.dislodgeds[prov]; ok {
		p = prov
		return
	}
	sup, _ := prov.Split()
	if u, ok = self.dislodgeds[sup]; ok {
		p = sup
		return
	}
	for _, name := range self.graph.Coasts(prov) {
		if u, ok = self.dislodgeds[name]; ok {
			p = name
			return
		}
	}
	return
}

func (self *State) Unit(prov common.Province) (u common.Unit, p common.Province, ok bool) {
	if u, ok = self.units[prov]; ok {
		p = prov
		return
	}
	sup, _ := prov.Split()
	if u, ok = self.units[sup]; ok {
		p = sup
		return
	}
	for _, name := range self.graph.Coasts(prov) {
		if u, ok = self.units[name]; ok {
			p = name
			return
		}
	}
	return
}

func (self *State) SupplyCenter(prov common.Province) (n common.Nation, p common.Province, ok bool) {
	if n, ok = self.supplyCenters[prov]; ok {
		p = prov
		return
	}
	sup, _ := prov.Split()
	if n, ok = self.supplyCenters[sup]; ok {
		p = sup
		return
	}
	for _, name := range self.graph.Coasts(prov) {
		if n, ok = self.supplyCenters[name]; ok {
			p = name
			return
		}
	}
	return
}

func (self *State) Order(prov common.Province) (o common.Order, p common.Province, ok bool) {
	if o, ok = self.orders[prov]; ok {
		p = prov
		return
	}
	sup, _ := prov.Split()
	if o, ok = self.orders[sup]; ok {
		p = sup
		return
	}
	for _, name := range self.graph.Coasts(prov) {
		if o, ok = self.orders[name]; ok {
			p = name
			return
		}
	}
	return
}

// Mutators

func (self *State) Move(src, dst common.Province, preventRetreat bool) {
	self.movements = append(self.movements, &movement{
		src:            src,
		dst:            dst,
		preventRetreat: preventRetreat,
	})
}

func (self *State) Retreat(src, dst common.Province) {
	if unit, prov, ok := self.Dislodged(src); !ok {
		panic(fmt.Errorf("No dislodged at %v?", src))
	} else {
		self.RemoveDislodged(prov)
		self.SetUnit(dst, unit)
		common.Logf("Moving dislodged %v from %v to %v", unit, src, dst)
	}
}
