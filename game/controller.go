package game

import (
	"encoding/base64"
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
	"net/url"
)

func DeleteMember(c common.Context, gameId, email string) error {
	return c.DB().Transact(func(d *kol.DB) error {
		urlDecodedId, err := url.QueryUnescape(gameId)
		if err != nil {
			return err
		}
		base64DecodedId, err := base64.StdEncoding.DecodeString(urlDecodedId)
		if err != nil {
			return err
		}
		game := Game{Id: base64DecodedId}
		if err := d.Get(&game); err != nil {
			return fmt.Errorf("Game not found: %v", err)
		}
		if game.Started {
			return fmt.Errorf("%+v already started", game)
		}
		member := Member{}
		if _, err := d.Query().Where(kol.And{kol.Equals{"GameId", base64DecodedId}, kol.Equals{"UserId", []byte(email)}}).First(&member); err != nil {
			return err
		}
		if err := d.Del(&member); err != nil {
			return err
		}
		left := game.Members(d)
		if len(left) == 0 {
			if err := d.Del(&game); err != nil {
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
		if game.Started {
			return fmt.Errorf("%+v already started")
		}
		variant, found := common.VariantMap[game.Variant]
		if !found {
			return fmt.Errorf("Unknown variant %v", game.Variant)
		}
		already := game.Members(d)
		if len(already) < len(variant.Nations) {
			id := make([]byte, len(state.Game.Id)+len([]byte(email)))
			copy(id, state.Game.Id)
			copy(id[len(state.Game.Id):], []byte(email))
			member := Member{
				Id:               id,
				GameId:           state.Game.Id,
				UserId:           []byte(email),
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
		UserId:           []byte(creator),
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
