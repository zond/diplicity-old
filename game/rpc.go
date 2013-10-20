package game

import (
	"encoding/base64"
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/godip/classical/orders"
	dip "github.com/zond/godip/common"
)

func SetOrder(c common.Context, gameId string, order []string, email string) (err error) {
	var base64DecodedId []byte
	if base64DecodedId, err = base64.URLEncoding.DecodeString(gameId); err != nil {
		return
	}
	game := Game{Id: base64DecodedId}
	if err = c.DB().Get(&game); err != nil {
		return
	}
	var member *Member
	if member, err = game.Member(c.DB(), email); err != nil {
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
	var parsedOrder dip.Order
	parsedOrder, err = orders.Parse(order)
	if err != nil {
		return
	}
	if err = parsedOrder.Validate(phase.GetState()); err != nil {
		return
	}
	nationOrders[dip.Province(order[0])] = order[1:]
	return c.DB().Set(phase)
}

func GetPossibleSources(c common.Context, gameId, email string) (result []dip.Province, err error) {
	var base64DecodedId []byte
	base64DecodedId, err = base64.URLEncoding.DecodeString(gameId)
	if err != nil {
		return
	}
	game := Game{Id: base64DecodedId}
	if err = c.DB().Get(&game); err != nil {
		return
	}
	var member *Member
	member, err = game.Member(c.DB(), email)
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

func GetValidOrders(c common.Context, gameId, province, email string) (result dip.Options, err error) {
	var base64DecodedId []byte
	base64DecodedId, err = base64.URLEncoding.DecodeString(gameId)
	if err != nil {
		return
	}
	game := Game{Id: base64DecodedId}
	if err = c.DB().Get(&game); err != nil {
		return
	}
	var member *Member
	member, err = game.Member(c.DB(), email)
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
	nation, options, found := phase.GetState().Options(orders.Types(), dip.Province(province))
	if found && nation == member.Nation {
		result = options
	}
	return
}
