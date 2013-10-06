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
	base64DecodedId, err = base64.StdEncoding.DecodeString(gameId)
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
	phase := game.LastPhase(c.DB())
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

func GetValidOrders(c common.Context, gameId, province, email string) (result dip.Options, err error) {
	var base64DecodedId []byte
	base64DecodedId, err = base64.StdEncoding.DecodeString(gameId)
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
	phase := game.LastPhase(c.DB())
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
