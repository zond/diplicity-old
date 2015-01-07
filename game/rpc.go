package game

import (
	"encoding/base64"
	"fmt"
	"github.com/zond/diplicity/srv"
	"github.com/zond/godip/classical/orders"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/state"
)

func UncommitPhase(c srv.WSContext) (result interface{}, err error) {
	err = setPhaseCommitted(c, false)
	return
}

func SeeMessage(c srv.WSContext) (result interface{}, err error) {
	messageId, err := base64.URLEncoding.DecodeString(c.Data().GetString("MessageId"))
	if err != nil {
		err = fmt.Errorf("Error trying to decode %#v: %v", c.Data().GetString("MessageId"), err)
		return
	}
	err = c.Update(func(c srv.WSTXContext) (err error) {
		message := &Message{Id: messageId}
		if err = c.TX().Get(message); err != nil {
			return
		}
		game := &Game{Id: message.GameId}
		if err = c.TX().Get(game); err != nil {
			return
		}
		var member *Member
		if member, err = game.Member(c.TX(), c.Principal()); err != nil {
			return
		}
		if message.SeenBy == nil {
			message.SeenBy = map[string]bool{}
		}
		if !message.SeenBy[member.Id.String()] {
			message.SeenBy[member.Id.String()] = true
			if err = c.TX().Set(message); err != nil {
				return
			}
		}
		return
	})
	return
}

func CommitPhase(c srv.WSContext) (result interface{}, err error) {
	err = setPhaseCommitted(c, true)
	return
}

func setPhaseCommitted(c srv.WSContext, commit bool) (err error) {
	phaseId, err := base64.URLEncoding.DecodeString(c.Data().GetString("PhaseId"))
	if err != nil {
		return
	}
	return c.Update(func(c srv.WSTXContext) (err error) {
		phase := &Phase{Id: phaseId}
		if err = c.TX().Get(phase); err != nil {
			return
		}
		game, err := phase.Game(c.TX())
		if err != nil {
			return
		}
		members, err := game.Members(c.TX())
		if err != nil {
			return
		}
		member := members.Get(c.Principal())
		if member == nil {
			err = fmt.Errorf("Not member of game")
			return
		}
		if member.NoOrders {
			c.Infof("%+v has no orders to give", member)
			err = fmt.Errorf("No orders to give")
			return
		}
		member.Committed = commit
		member.NoWait = false
		if err = c.TX().Set(member); err != nil {
			return
		}
		if !phase.Resolved {
			count := 0
			for _, m := range members {
				if m.Committed || m.NoWait {
					count++
				}
			}
			if count == len(members) {
				if err = game.resolve(c.TXDiet(), phase); err != nil {
					return
				}
				c.Infof("Resolved %v", game.Id)
				return
			}
		}
		err = c.TX().Set(phase)
		return
	})
}

func SetOrder(c srv.WSContext) (result interface{}, err error) {
	var base64DecodedId []byte
	if base64DecodedId, err = base64.URLEncoding.DecodeString(c.Data().GetString("GameId")); err != nil {
		return
	}
	err = c.Update(func(c srv.WSTXContext) (err error) {
		game := Game{Id: base64DecodedId}
		if err = c.TX().Get(&game); err != nil {
			return
		}
		var member *Member
		if member, err = game.Member(c.TX(), c.Principal()); err != nil {
			return
		}
		var phase *Phase
		if _, phase, err = game.Phase(c.TX(), 0); err != nil {
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
		if err = c.TX().Set(phase); err != nil {
			return
		}
		return
	})
	return
}
