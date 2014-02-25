package game

import (
	"encoding/base64"
	"fmt"

	"github.com/zond/godip/classical/orders"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/subs"
)

func UncommitPhase(c subs.Context) (result interface{}, err error) {
	err = setPhaseCommitted(c, false)
	return
}

func CommitPhase(c subs.Context) (result interface{}, err error) {
	err = setPhaseCommitted(c, true)
	return
}

func setPhaseCommitted(c subs.Context, commit bool) (err error) {
	phaseId, err := base64.URLEncoding.DecodeString(c.Data().GetString("PhaseId"))
	if err != nil {
		return
	}
	phase := Phase{Id: phaseId}
	if err = c.DB().Get(&phase); err != nil {
		return
	}
	game, err := phase.Game(c.DB())
	if err != nil {
		return
	}
	member, err := game.Member(c.DB(), c.Principal())
	if err != nil {
		return
	}
	phase.Committed[member.Nation] = commit
	err = c.DB().Set(&phase)
	return
}

func SetOrder(c subs.Context) (result interface{}, err error) {
	var base64DecodedId []byte
	if base64DecodedId, err = base64.URLEncoding.DecodeString(c.Data().GetString("GameId")); err != nil {
		return
	}
	game := Game{Id: base64DecodedId}
	if err = c.DB().Get(&game); err != nil {
		return
	}
	var member *Member
	if member, err = game.Member(c.DB(), c.Principal()); err != nil {
		return
	}
	var phase *Phase
	if phase, err = game.LastPhase(c.DB()); err != nil {
		return
	}
	if phase == nil {
		err = fmt.Errorf("No phase for %+v found", game)
		return
	}
	if phase.Orders == nil {
		phase.Orders = map[dip.Nation]map[dip.Province][]string{}
	}
	nationOrders, found := phase.Orders[member.Nation]
	if !found {
		nationOrders = map[dip.Province][]string{}
		phase.Orders[member.Nation] = nationOrders
	}
	order := c.Data().GetStringSlice("Order")
	var parsedOrder dip.Order
	parsedOrder, err = orders.Parse(order)
	if err != nil {
		return
	}
	if err = parsedOrder.Validate(phase.GetState()); err != nil {
		return
	}
	nationOrders[dip.Province(order[0])] = order[1:]
	if err = c.DB().Set(phase); err != nil {
		return
	}
	return
}

func GetPossibleSources(c subs.Context) (result interface{}, err error) {
	var base64DecodedId []byte
	base64DecodedId, err = base64.URLEncoding.DecodeString(c.Data().GetString("GameId"))
	if err != nil {
		return
	}
	game := Game{Id: base64DecodedId}
	if err = c.DB().Get(&game); err != nil {
		return
	}
	var member *Member
	member, err = game.Member(c.DB(), c.Principal())
	if err != nil {
		return
	}
	var phase *Phase
	if phase, err = game.LastPhase(c.DB()); err != nil {
		return
	}
	if phase == nil {
		err = fmt.Errorf("No phase for %+v found", game)
		return
	}
	state := phase.GetState()
	result = state.Phase().PossibleSources(state, member.Nation)
	return
}

func GetValidOrders(c subs.Context) (result interface{}, err error) {
	var base64DecodedId []byte
	base64DecodedId, err = base64.URLEncoding.DecodeString(c.Data().GetString("GameId"))
	if err != nil {
		return
	}
	game := Game{Id: base64DecodedId}
	if err = c.DB().Get(&game); err != nil {
		return
	}
	var member *Member
	member, err = game.Member(c.DB(), c.Principal())
	if err != nil {
		return
	}
	var phase *Phase
	if phase, err = game.LastPhase(c.DB()); err != nil {
		return
	}
	if phase == nil {
		err = fmt.Errorf("No phase for %+v found", game)
		return
	}
	nation, options, found := phase.GetState().Options(orders.Types(), dip.Province(c.Data().GetString("Province")))
	if found && nation == member.Nation {
		result = options
	}
	return
}
