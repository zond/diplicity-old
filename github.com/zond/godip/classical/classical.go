package classical

import (
	"fmt"
	cla "github.com/zond/godip/classical/common"
	"github.com/zond/godip/classical/start"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/state"
)

func Blank(phase dip.Phase) *state.State {
	return state.New(start.Graph(), phase, BackupRule)
}

func Start() *state.State {
	return state.New(start.Graph(), &phase{1901, cla.Spring, cla.Movement}, BackupRule).
		SetUnits(start.Units()).
		SetSupplyCenters(start.SupplyCenters())
}

func BackupRule(state dip.State, deps []dip.Province) {
	only_moves := true
	convoys := false
	for _, prov := range deps {
		if order, _, ok := state.Order(prov); ok {
			if order.Type() != cla.Move {
				only_moves = false
			}
			if order.Type() == cla.Convoy {
				convoys = true
			}
		}
	}

	if only_moves {
		for _, prov := range deps {
			state.SetResolution(prov, nil)
		}
		return
	}
	if convoys {
		for _, prov := range deps {
			if order, _, ok := state.Order(prov); ok && order.Type() == cla.Convoy {
				state.SetResolution(prov, cla.ErrConvoyParadox)
			}
		}
		return
	}

	panic(fmt.Errorf("Unknown circular dependency between %v", deps))
}
