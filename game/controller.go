package game

import (
	"fmt"
	dip "github.com/zond/godip/common"

	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
)

func CreateMessage(c common.Context, j subs.JSON, senderEmail string) (err error) {
	// load the  message provided by the client
	var message Message
	j.Overwrite(&message)

	if message.Body == "" {
		return
	}

	// and the game
	game := &Game{Id: message.GameId}
	if err := c.DB().Get(game); err != nil {
		return err
	}
	// and the member
	sender, err := game.Member(c.DB(), senderEmail)
	if err != nil {
		return
	}

	// make sure the sender is one of the recipients
	message.Recipients[sender.Id.String()] = true

	var phaseType dip.PhaseType
	switch game.State {
	case common.GameStateCreated:
		phaseType = common.BeforeGamePhaseType
	case common.GameStateStarted:
		var phase *Phase
		if phase, err = game.LastPhase(c.DB()); err != nil {
			return
		}
		phaseType = phase.Type
	case common.GameStateEnded:
		phaseType = common.AfterGamePhaseType
	default:
		err = fmt.Errorf("Unknown game state for %+v", game)
		return
	}

	allowedFlags := game.ChatFlags[phaseType]

	recipients := len(message.Recipients)
	if recipients == 2 {
		if (allowedFlags & common.ChatPrivate) == 0 {
			err = fmt.Errorf("%+v does not allow %+v during %+v (%v)", game, message, phaseType, common.ChatPrivate)
			return
		}
	} else if recipients == len(common.VariantMap[game.Variant].Nations) {
		if (allowedFlags & common.ChatConference) == 0 {
			err = fmt.Errorf("%+v does not allow %+v during %+v", game, message, phaseType)
			return
		}
	} else if recipients > 2 {
		if (allowedFlags & common.ChatGroup) == 0 {
			err = fmt.Errorf("%+v does not allow %+v during %+v", game, message, phaseType)
			return
		}
	} else {
		err = fmt.Errorf("%+v doesn't have any recipients", message)
		return
	}

	if err = c.DB().Set(&message); err != nil {
		return
	}

	return
}

func DeleteMember(c common.Context, gameId, email string) error {
	return c.DB().Transact(func(d *kol.DB) error {
		decodedId, err := kol.DecodeId(gameId)
		if err != nil {
			return err
		}
		game := &Game{Id: decodedId}
		if err := d.Get(game); err != nil {
			return fmt.Errorf("Game not found: %v", err)
		}
		if game.State != common.GameStateCreated {
			return fmt.Errorf("%+v already started", game)
		}
		member := Member{}
		if _, err := d.Query().Where(kol.And{kol.Equals{"GameId", decodedId}, kol.Equals{"UserId", kol.Id(email)}}).First(&member); err != nil {
			return err
		}
		if err := d.Del(&member); err != nil {
			return err
		}
		left, err := game.Members(d)
		if err != nil {
			return err
		}
		if len(left) == 0 {
			if err := d.Del(game); err != nil {
				return err
			}
		}
		return nil
	})
}

func AddMember(c common.Context, j subs.JSON, email string) error {
	var state GameState
	j.Overwrite(&state)
	return c.DB().Transact(func(d *kol.DB) error {
		game := Game{Id: state.Game.Id}
		if err := d.Get(&game); err != nil {
			return fmt.Errorf("Game not found")
		}
		if game.State != common.GameStateCreated {
			return fmt.Errorf("%+v already started")
		}
		variant, found := common.VariantMap[game.Variant]
		if !found {
			return fmt.Errorf("Unknown variant %v", game.Variant)
		}
		if alreadyMember, err := game.Member(d, email); err != nil {
			return err
		} else if alreadyMember != nil {
			return fmt.Errorf("%+v is already member of %v", alreadyMember, game.Id)
		}
		me := user.EnsureUser(d, email)
		if game.Disallows(me) {
			return fmt.Errorf("Is not allowed to join this game due to game settings")
		}
		already, err := game.Members(d)
		if err != nil {
			return err
		}
		if disallows, err := already.Disallows(d, me); err != nil {
			return err
		} else if disallows {
			return fmt.Errorf("Is not allowed to join this game due to blacklistings")
		}
		if len(already) < len(variant.Nations) {
			member := Member{
				GameId:           state.Game.Id,
				UserId:           kol.Id(email),
				PreferredNations: state.Members[0].PreferredNations,
			}
			if err := d.Set(&member); err != nil {
				return err
			}
			if len(already) == len(variant.Nations)-1 {
				if err := game.start(d); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func Create(c common.Context, j subs.JSON, creator string) error {
	var state GameState
	j.Overwrite(&state)

	game := &Game{
		Variant:          state.Game.Variant,
		EndYear:          state.Game.EndYear,
		Private:          state.Game.Private,
		SecretEmail:      state.Game.SecretEmail,
		SecretNickname:   state.Game.SecretNickname,
		SecretNation:     state.Game.SecretNation,
		Deadlines:        state.Game.Deadlines,
		ChatFlags:        state.Game.ChatFlags,
		AllocationMethod: state.Game.AllocationMethod,
	}

	if _, found := common.VariantMap[game.Variant]; !found {
		return fmt.Errorf("Unknown variant for %+v", game)
	}

	if _, found := common.AllocationMethodMap[game.AllocationMethod]; !found {
		return fmt.Errorf("Unknown allocation method for %+v", game)
	}

	member := &Member{
		UserId:           kol.Id(creator),
		PreferredNations: state.Members[0].PreferredNations,
	}
	return c.DB().Transact(func(d *kol.DB) error {
		if err := d.Set(game); err != nil {
			return err
		}
		member.GameId = game.Id
		return d.Set(member)
	})
}
