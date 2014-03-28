package game

import (
	"encoding/base64"
	"fmt"

	"github.com/zond/diplicity/common"
	"github.com/zond/godip/classical/orders"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/state"
	"github.com/zond/kcwraps/kol"
)

func UncommitPhase(c common.WSContext) (result interface{}, err error) {
	err = setPhaseCommitted(c, false)
	return
}

func SeeMessage(c common.WSContext) (result interface{}, err error) {
	messageId, err := base64.URLEncoding.DecodeString(c.Data().GetString("MessageId"))
	if err != nil {
		return
	}
	err = c.Transact(func(c common.WSContext) (err error) {
		message := &Message{Id: messageId}
		if err = c.DB().Get(message); err != nil {
			return
		}
		game := &Game{Id: message.GameId}
		if err = c.DB().Get(game); err != nil {
			return
		}
		var member *Member
		if member, err = game.Member(c.DB(), c.Principal()); err != nil {
			return
		}
		if message.SeenBy == nil {
			message.SeenBy = map[string]bool{}
		}
		if !message.SeenBy[member.Id.String()] {
			message.SeenBy[member.Id.String()] = true
			if err = c.DB().Set(message); err != nil {
				return
			}
		}
		return
	})
	return
}

func CommitPhase(c common.WSContext) (result interface{}, err error) {
	err = setPhaseCommitted(c, true)
	return
}

func setPhaseCommitted(c common.WSContext, commit bool) (err error) {
	phaseId, err := base64.URLEncoding.DecodeString(c.Data().GetString("PhaseId"))
	if err != nil {
		return
	}
	return c.Transact(func(c common.WSContext) (err error) {
		phase := &Phase{Id: phaseId}
		if err = c.DB().Get(phase); err != nil {
			return
		}
		game, err := phase.Game(c.DB())
		if err != nil {
			return
		}
		members, err := game.Members(c.DB())
		if err != nil {
			return
		}
		member := members.Get(c.Principal())
		if member == nil {
			err = fmt.Errorf("Not member of game")
			return
		}
		member.Committed = commit
		if err = c.DB().Set(member); err != nil {
			return
		}
		if !phase.Resolved {
			count := 0
			for _, m := range members {
				if m.Committed {
					count++
				}
			}
			if count == len(members) {
				if err = game.resolve(c.Diet(), phase); err != nil {
					return
				}
				c.Infof("Resolved %v", game.Id)
				return
			}
		}
		err = c.DB().Set(phase)
		return
	})
}

func SetOrder(c common.WSContext) (result interface{}, err error) {
	var base64DecodedId []byte
	if base64DecodedId, err = base64.URLEncoding.DecodeString(c.Data().GetString("GameId")); err != nil {
		return
	}
	err = c.DB().Transact(func(d *kol.DB) (err error) {
		game := Game{Id: base64DecodedId}
		if err = d.Get(&game); err != nil {
			return
		}
		var member *Member
		if member, err = game.Member(d, c.Principal()); err != nil {
			return
		}
		var phase *Phase
		if _, phase, err = game.Phase(d, 0); err != nil {
			return
		}
		if phase == nil {
			err = fmt.Errorf("No phase for %+v found", game)
			return
		}
		nationOrders, found := phase.Orders[member.Nation]
		if !found {
			nationOrders = map[dip.Province][]string{}
			phase.Orders[member.Nation] = nationOrders
		}
		order := c.Data().GetStringSlice("Order")
		if len(order) == 1 {
			delete(nationOrders, dip.Province(order[0]))
		} else {
			var parsedOrder dip.Order
			parsedOrder, err = orders.Parse(order)
			if err != nil {
				return
			}
			var state *state.State
			state, err = phase.State()
			if err != nil {
				return
			}
			if err = parsedOrder.Validate(state); err != nil {
				return
			}
			nationOrders[dip.Province(order[0])] = order[1:]
		}
		if err = d.Set(phase); err != nil {
			return
		}
		return
	})
	return
}
