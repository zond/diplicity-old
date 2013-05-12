package common

import (
	"fmt"
	. "github.com/zond/godip/common"
	"sort"
)

const (
	Sea  Flag = "Sea"
	Land Flag = "Land"

	Army  UnitType = "Army"
	Fleet UnitType = "Fleet"

	England Nation = "England"
	France  Nation = "France"
	Germany Nation = "Germany"
	Russia  Nation = "Russia"
	Austria Nation = "Austria"
	Italy   Nation = "Italy"
	Turkey  Nation = "Turkey"
	Neutral Nation = "Neutral"

	Spring Season = "Spring"
	Fall   Season = "Fall"

	Movement   PhaseType = "Movement"
	Retreat    PhaseType = "Retreat"
	Adjustment PhaseType = "Adjustment"

	Build   OrderType = "Build"
	Move    OrderType = "Move"
	Hold    OrderType = "Hold"
	Convoy  OrderType = "Convoy"
	Support OrderType = "Support"
	Disband OrderType = "Disband"

	ViaConvoy Flag = "C"
)

var Coast = []Flag{Sea, Land}

var Nations = []Nation{Austria, England, France, Germany, Italy, Turkey, Russia}
var PhaseTypes = []PhaseType{Movement, Retreat, Adjustment}

// Invalid is not understood
// Illegal is understood but not allowed
var ErrInvalidSource = fmt.Errorf("ErrInvalidSource")
var ErrInvalidDestination = fmt.Errorf("ErrInvalidDestination")
var ErrInvalidTarget = fmt.Errorf("ErrInvalidTarget")
var ErrInvalidPhase = fmt.Errorf("ErrInvalidPhase")
var ErrMissingUnit = fmt.Errorf("ErrMissingUnit")
var ErrIllegalDestination = fmt.Errorf("ErrIllegalDestination")
var ErrMissingConvoyPath = fmt.Errorf("ErrMissignConvoyPath")
var ErrIllegalMove = fmt.Errorf("ErrIllegalMove")
var ErrConvoyParadox = fmt.Errorf("ErrConvoyParadox")
var ErrIllegalSupportPosition = fmt.Errorf("ErrIllegalSupportPosition")
var ErrIllegalSupportDestination = fmt.Errorf("ErrIllegalSupportDestination")
var ErrIllegalSupportDestinationNation = fmt.Errorf("ErrIllegalSupportDestinationNation")
var ErrMissingSupportUnit = fmt.Errorf("ErrMissingSupportUnit")
var ErrIllegalSupportMove = fmt.Errorf("ErrIllegalSupportMove")
var ErrIllegalConvoyUnit = fmt.Errorf("ErrIllegalConvoyUnit")
var ErrIllegalConvoyMove = fmt.Errorf("ErrIllegalConvoyMove")
var ErrMissingConvoyee = fmt.Errorf("ErrMissingConvoyee")
var ErrIllegalBuild = fmt.Errorf("ErrIllegalBuild")
var ErrIllegalDisband = fmt.Errorf("ErrIllegalDisband")
var ErrOccupiedSupplyCenter = fmt.Errorf("ErrOccupiedSupplyCenter")
var ErrMissingSupplyCenter = fmt.Errorf("ErrMissingSupplyCenter")
var ErrMissingSurplus = fmt.Errorf("ErrMissingSurplus")
var ErrIllegalUnitType = fmt.Errorf("ErrIllegalUnitType")
var ErrMissingDeficit = fmt.Errorf("ErrMissingDeficit")
var ErrOccupiedDestination = fmt.Errorf("ErrOccupiedDestination")
var ErrIllegalRetreat = fmt.Errorf("ErrIllegalRetreat")
var ErrForcedDisband = fmt.Errorf("ErrForcedDisband")
var ErrHostileSupplyCenter = fmt.Errorf("ErrHostileSupplyCenter")

type ErrConvoyDislodged struct {
	Province Province
}

func (self ErrConvoyDislodged) Error() string {
	return fmt.Sprintf("ErrConvoyDislodged:%v", self.Province)
}

type ErrSupportBroken struct {
	Province Province
}

func (self ErrSupportBroken) Error() string {
	return fmt.Sprintf("ErrSupportBroken:%v", self.Province)
}

type ErrBounce struct {
	Province Province
}

func (self ErrBounce) Error() string {
	return fmt.Sprintf("ErrBounce:%v", self.Province)
}

func convoyPath(v Validator, src, dst Province, resolveConvoys bool, viaNation *Nation) []Province {
	if src == dst {
		return nil
	}
	waypoints, _, _ := v.Find(func(p Province, o Order, u *Unit) bool {
		if !v.Graph().Flags(p)[Land] && u != nil && (viaNation == nil || u.Nation == *viaNation) && u.Type == Fleet {
			if !resolveConvoys {
				if viaNation == nil || (o != nil && o.Type() == Convoy && o.Targets()[1].Contains(src) && o.Targets()[2].Contains(dst)) {
					return true
				}
				return false
			}
			if o != nil && o.Type() == Convoy && o.Targets()[1].Contains(src) && o.Targets()[2].Contains(dst) {
				if err := v.(Resolver).Resolve(p); err != nil {
					return false
				}
				return true
			}
		}
		return false
	})
	for _, waypoint := range waypoints {
		filter := func(name Province, edgeFlags, nodeFlags map[Flag]bool, sc *Nation) bool {
			if name.Contains(dst) {
				return true
			}
			if nodeFlags[Land] {
				return false
			}
			if u, _, ok := v.Unit(name); ok && u.Type == Fleet {
				if !resolveConvoys {
					return true
				}
				if order, prov, ok := v.Order(name); ok && order.Type() == Convoy && order.Targets()[1].Contains(src) && order.Targets()[2].Contains(dst) {
					if err := v.(Resolver).Resolve(prov); err != nil {
						return false
					}
					return true
				}
			}
			return false
		}
		if part1 := v.Graph().Path(src, waypoint, filter); part1 != nil {
			if part2 := v.Graph().Path(waypoint, dst, filter); part2 != nil {
				return append(part1, part2...)
			}
		}
	}
	return nil
}

func Path(v Validator, typ UnitType, src, dst Province) []Province {
	var filter PathFilter
	if typ == Army {
		filter = func(p Province, ef, nf map[Flag]bool, sc *Nation) bool {
			return ef[Land] && nf[Land]
		}
	} else {
		filter = func(p Province, ef, nf map[Flag]bool, sc *Nation) bool {
			return ef[Sea] && nf[Sea]
		}
	}
	return v.Graph().Path(src, dst, filter)
}

func MustConvoy(r Resolver, src Province) bool {
	unit, _, ok := r.Unit(src)
	if !ok {
		return false
	}
	if unit.Type != Army {
		return false
	}
	order, _, ok := r.Order(src)
	if !ok {
		return false
	}
	if order.Type() != Move {
		return false
	}
	path := Path(r, unit.Type, order.Targets()[0], order.Targets()[1])
	return (path == nil ||
		len(path) > 1 ||
		(order.Flags()[ViaConvoy] && AnyConvoyPath(r, order.Targets()[0], order.Targets()[1], true, nil) != nil) ||
		AnyConvoyPath(r, order.Targets()[0], order.Targets()[1], false, &unit.Nation) != nil)
}

func AnyConvoyPath(v Validator, src, dst Province, resolveConvoys bool, viaNation *Nation) (result []Province) {
	if result = convoyPath(v, src, dst, resolveConvoys, viaNation); result != nil {
		return
	}
	for _, srcCoast := range v.Graph().Coasts(src) {
		for _, dstCoast := range v.Graph().Coasts(dst) {
			if result = convoyPath(v, srcCoast, dstCoast, resolveConvoys, viaNation); result != nil {
				return
			}
		}
	}
	return
}

func AnySupportPossible(v Validator, typ UnitType, src, dst Province) (err error) {
	if err = movePossible(v, typ, src, dst, false, false); err == nil {
		return
	}
	for _, coast := range v.Graph().Coasts(dst) {
		if err = movePossible(v, typ, src, coast, false, false); err == nil {
			return
		}
	}
	return
}

func AnyMovePossible(v Validator, typ UnitType, src, dst Province, lax, allowConvoy, resolveConvoys bool) (dstCoast Province, err error) {
	dstCoast = dst
	if err = movePossible(v, typ, src, dst, allowConvoy, resolveConvoys); err == nil {
		return
	}
	if lax || dst.Super() == dst {
		var options []Province
		for _, coast := range v.Graph().Coasts(dst) {
			if err2 := movePossible(v, typ, src, coast, allowConvoy, resolveConvoys); err2 == nil {
				options = append(options, coast)
			}
		}
		if len(options) > 0 {
			if lax || len(options) == 1 {
				dstCoast, err = options[0], nil
			}
		}
	}
	return
}

func movePossible(v Validator, typ UnitType, src, dst Province, allowConvoy, resolveConvoys bool) error {
	if !v.Graph().Has(src) {
		return ErrInvalidSource
	}
	if !v.Graph().Has(dst) {
		return ErrInvalidDestination
	}
	if typ == Army {
		if !v.Graph().Flags(dst)[Land] {
			return ErrIllegalDestination
		}
		if allowConvoy && resolveConvoys {
			if MustConvoy(v.(Resolver), src) {
				if AnyConvoyPath(v, src, dst, true, nil) == nil {
					return ErrMissingConvoyPath
				}
				return nil
			}
			return nil
		}
		if path := Path(v, typ, src, dst); path == nil || len(path) > 1 {
			if allowConvoy {
				if cp := AnyConvoyPath(v, src, dst, false, nil); cp == nil {
					return ErrMissingConvoyPath
				}
				return nil
			}
			return ErrIllegalMove
		}
		return nil
	} else if typ == Fleet {
		if !v.Graph().Flags(dst)[Sea] {
			return ErrIllegalDestination
		}
		if path := Path(v, typ, src, dst); path == nil || len(path) > 1 {
			return ErrIllegalMove
		}
	}
	return nil
}

func AdjustmentStatus(v Validator, me Nation) (builds Orders, disbands Orders, balance int) {
	scs := 0
	for prov, nat := range v.SupplyCenters() {
		if nat == me {
			scs += 1
			if order, _, ok := v.Order(prov); ok && order.Type() == Build {
				builds = append(builds, order)
			}
		}
	}

	units := 0
	for prov, unit := range v.Units() {
		if unit.Nation == me {
			units += 1
			if order, _, ok := v.Order(prov); ok && order.Type() == Disband {
				disbands = append(disbands, order)
			}
		}
	}

	sort.Sort(builds)
	sort.Sort(disbands)

	balance = scs - units
	if balance > 0 {
		disbands = nil
		builds = builds[:Max(0, Min(len(builds), balance))]
	} else if balance < 0 {
		builds = nil
		disbands = disbands[:Max(0, Min(len(disbands), -balance))]
	} else {
		builds = nil
		disbands = nil
	}

	return
}

/*
HoldSupport returns successful supports of a hold in prov.
*/
func HoldSupport(r Resolver, prov Province) int {
	_, supports, _ := r.Find(func(p Province, o Order, u *Unit) bool {
		if o != nil && u != nil && o.Type() == Support && p.Super() != prov.Super() && len(o.Targets()) == 2 && o.Targets()[1].Super() == prov.Super() {
			if err := r.Resolve(p); err == nil {
				return true
			}
		}
		return false
	})
	return len(supports)
}

/*
MoveSupport returns the successful supports of movement from src to dst, discounting the nations in forbiddenSupports.
*/
func MoveSupport(r Resolver, src, dst Province, forbiddenSupports []Nation) int {
	_, supports, _ := r.Find(func(p Province, o Order, u *Unit) bool {
		if o != nil && u != nil {
			if o.Type() == Support && len(o.Targets()) == 3 && o.Targets()[1].Contains(src) && o.Targets()[2].Contains(dst) {
				for _, ban := range forbiddenSupports {
					if ban == u.Nation {
						return false
					}
				}
				if err := r.Resolve(p); err == nil {
					return true
				}
			}
		}
		return false
	})
	return len(supports)
}
