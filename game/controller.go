package game

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/epoch"
	"github.com/zond/diplicity/game/allocation"
	"github.com/zond/diplicity/game/meta"
	"github.com/zond/diplicity/user"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/variants"
	"github.com/zond/kcwraps/kol"
)

func Resolve(c *common.HTTPContext) (err error) {
	phase := &Phase{}
	if err = json.NewDecoder(c.Req().Body).Decode(phase); err != nil {
		return
	}
	state, err := phase.State()
	if err != nil {
		return
	}
	if err = state.Next(); err != nil {
		return
	}
	// Load the new godip phase from the state
	nextDipPhase := state.Phase()
	// Create a diplicity phase for the new phase
	nextPhase := &Phase{
		Ordinal:     phase.Ordinal + 1,
		Orders:      map[dip.Nation]map[dip.Province][]string{},
		Resolutions: map[dip.Province]string{},
		Season:      nextDipPhase.Season(),
		Year:        nextDipPhase.Year(),
		Type:        nextDipPhase.Type(),
	}
	// Set the new phase positions
	var resolutions map[dip.Province]error
	nextPhase.Units, nextPhase.SupplyCenters, nextPhase.Dislodgeds, nextPhase.Dislodgers, nextPhase.Bounces, resolutions = state.Dump()
	for prov, err := range resolutions {
		if err == nil {
			nextPhase.Resolutions[prov] = "OK"
		} else {
			nextPhase.Resolutions[prov] = err.Error()
		}
	}
	c.Resp().Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err = json.NewEncoder(c.Resp()).Encode(nextPhase); err != nil {
		return
	}
	return
}

func UnsubscribeEmails(c *common.HTTPContext) (err error) {
	unsubTag, err := common.DecodeUnsubscribeTag(c.Secret(), c.Vars()["unsubscribe_tag"])
	if err != nil {
		return
	}
	u := &user.User{Id: unsubTag.U}
	if err = c.DB().Get(u); err != nil {
		return
	}
	switch unsubTag.T {
	case common.UnsubscribeMessageEmail:
		u.MessageEmailDisabled = true
	case common.UnsubscribePhaseEmail:
		u.MessageEmailDisabled = true
	}
	if err = c.DB().Set(u); err != nil {
		return
	}
	switch unsubTag.T {
	case common.UnsubscribeMessageEmail:
		fmt.Fprintf(c.Resp(), "%v has successfully been unsubscribed from message emails.", u.Email)
	case common.UnsubscribePhaseEmail:
		fmt.Fprintf(c.Resp(), "%v has successfully been unsubscribed from phase emails.", u.Email)
	}
	return
}

func AdminGetOptions(c *common.HTTPContext) (err error) {
	gameId, err := base64.URLEncoding.DecodeString(c.Vars()["game_id"])
	if err != nil {
		return
	}
	game := &Game{Id: gameId}
	if err = c.DB().Get(game); err != nil {
		return
	}
	_, last, err := game.Phase(c.DB(), 0)
	if err != nil {
		return
	}
	opts, err := last.Options(dip.Nation(c.Vars()["nation"]))
	if err != nil {
		return
	}
	return c.RenderJSON(opts)
}

func AdminReindexGames(c *common.HTTPContext) (err error) {
	games := Games{}
	if err = c.DB().Query().All(&games); err != nil {
		return
	}
	for _, game := range games {
		if err = c.DB().Index(game); err != nil {
			return
		}
		fmt.Fprintf(c.Resp(), "Reindexed %#v\n", game.Id.String())
	}
	return
}

func AdminRecalcOptions(c *common.HTTPContext) (err error) {
	gameId, err := base64.URLEncoding.DecodeString(c.Vars()["game_id"])
	if err != nil {
		return
	}
	g := &Game{Id: gameId}
	if err = c.DB().Get(g); err != nil {
		return
	}
	_, last, err := g.Phase(c.DB(), 0)
	if err != nil {
		return
	}
	members, err := g.Members(c.DB())
	if err != nil {
		return
	}
	for index, _ := range members {
		opts := dip.Options{}
		if opts, err = last.Options(members[index].Nation); err != nil {
			return
		}
		members[index].Options = opts
		if len(opts) == 0 {
			members[index].Committed = true
			members[index].NoOrders = true
		} else {
			members[index].Committed = false
			members[index].NoOrders = false
		}
		if err = c.DB().Set(&members[index]); err != nil {
			return
		}
	}
	return
}

func AdminRollback(c *common.HTTPContext) (err error) {
	gameId, err := base64.URLEncoding.DecodeString(c.Vars()["game_id"])
	if err != nil {
		return
	}
	epoch, err := epoch.Get(c.DB())
	if err != nil {
		return
	}
	g := &Game{Id: gameId}
	if err = c.DB().Get(g); err != nil {
		return
	}
	ordinal, err := strconv.Atoi(c.Vars()["until"])
	if err != nil {
		return
	}
	members, err := g.Members(c.DB())
	if err != nil {
		return
	}
	phases, err := g.Phases(c.DB())
	if err != nil {
		return
	}
	sort.Sort(phases)
	for index, _ := range phases {
		phase := &phases[index]
		if phase.Ordinal == ordinal {
			phase.Resolutions = map[dip.Province]string{}
			phase.Resolved = false
			phase.Deadline = epoch + (time.Minute * time.Duration(g.Deadlines[phase.Type]))
			for index, _ := range members {
				opts := dip.Options{}
				if opts, err = phase.Options(members[index].Nation); err != nil {
					return
				}
				members[index].Options = opts
				if len(opts) == 0 {
					members[index].Committed = true
					members[index].NoOrders = true
				} else {
					members[index].Committed = false
					members[index].NoOrders = false
				}
				if err = c.DB().Set(&members[index]); err != nil {
					return
				}
			}
			if err = c.DB().Set(phase); err != nil {
				return
			}
		} else if phase.Ordinal > ordinal {
			if err = c.DB().Del(phase); err != nil {
				return
			}
		}
	}
	return
}

type AdminGameState struct {
	Game    *Game
	Phases  Phases
	Members []MemberState
}

func AdminGetGame(c *common.HTTPContext) (err error) {
	gameId, err := base64.URLEncoding.DecodeString(c.Vars()["game_id"])
	if err != nil {
		return
	}
	g := &Game{Id: gameId}
	if err = c.DB().Get(g); err != nil {
		return
	}
	members, err := g.Members(c.DB())
	if err != nil {
		return
	}
	memberStates, err := members.ToStates(c.DB(), g, "", true)
	if err != nil {
		return
	}
	phases, err := g.Phases(c.DB())
	if err != nil {
		return
	}
	sort.Sort(phases)
	return c.RenderJSON(AdminGameState{
		Game:    g,
		Phases:  phases,
		Members: memberStates,
	})
}

func CreateMessage(c common.WSContext) (err error) {
	// load the  message provided by the client
	message := &Message{}
	c.Data().Overwrite(message)
	if message.RecipientIds == nil {
		message.RecipientIds = map[string]bool{}
	}

	// set the body
	message.Body = strings.TrimSpace(message.Body)
	if message.Body == "" {
		return
	}

	// and the game
	game := &Game{Id: message.GameId}
	if err := c.DB().Get(game); err != nil {
		return err
	}
	// and the member
	sender, err := game.Member(c.DB(), c.Principal())
	if err != nil {
		return
	}

	return message.Send(c.Diet(), game, sender)
}

func DeleteMember(c common.WSContext) error {
	return c.Transact(func(c common.WSContext) error {
		decodedId, err := kol.DecodeId(c.Match()[1])
		if err != nil {
			return err
		}
		game := &Game{Id: decodedId}
		if err := c.DB().Get(game); err != nil {
			return fmt.Errorf("Game not found: %v", err)
		}
		if game.State != meta.GameStateCreated {
			return fmt.Errorf("%+v already started", game)
		}
		member := Member{}
		if _, err := c.DB().Query().Where(kol.And{kol.Equals{"GameId", decodedId}, kol.Equals{"UserId", kol.Id(c.Principal())}}).First(&member); err != nil {
			return err
		}
		if err := c.DB().Del(&member); err != nil {
			return err
		}
		left, err := game.Members(c.DB())
		if err != nil {
			return err
		}
		if len(left) == 0 {
			if err := c.DB().Del(game); err != nil {
				return err
			}
		}
		return nil
	})
}

func AddMember(c common.WSContext) error {
	var state GameState
	c.Data().Overwrite(&state)
	return c.Transact(func(c common.WSContext) error {
		game := Game{Id: state.Game.Id}
		if err := c.DB().Get(&game); err != nil {
			return fmt.Errorf("Game not found")
		}
		if game.State != meta.GameStateCreated {
			return fmt.Errorf("%+v already started")
		}
		variant, found := variants.Variants[game.Variant]
		if !found {
			return fmt.Errorf("Unknown variant %v", game.Variant)
		}
		if alreadyMember, err := game.Member(c.DB(), c.Principal()); err != nil {
			return err
		} else if alreadyMember != nil {
			return fmt.Errorf("%+v is already member of %v", alreadyMember, game.Id)
		}
		me := &user.User{Id: kol.Id(c.Principal())}
		if err := c.DB().Get(me); err != nil {
			return err
		}
		if game.Disallows(me) {
			return fmt.Errorf("Is not allowed to join this game due to game settings")
		}
		already, err := game.Members(c.DB())
		if err != nil {
			return err
		}
		if disallows, err := already.Disallows(c.DB(), me); err != nil {
			return err
		} else if disallows {
			return fmt.Errorf("Is not allowed to join this game due to blacklistings")
		}
		if len(already) < len(variant.Nations) {
			member := Member{
				GameId:           state.Game.Id,
				UserId:           kol.Id(c.Principal()),
				PreferredNations: state.Members[0].PreferredNations,
			}
			if err := c.DB().Set(&member); err != nil {
				return err
			}
			if len(already) == len(variant.Nations)-1 {
				if err := game.start(c.Diet()); err != nil {
					return err
				}
				c.Infof("Started %v", game.Id)
			}
		}
		return nil
	})
}

func Create(c common.WSContext) error {
	var state GameState
	c.Data().Overwrite(&state)

	game := &Game{
		Variant:               state.Game.Variant,
		EndYear:               state.Game.EndYear,
		Private:               state.Game.Private,
		SecretEmail:           state.Game.SecretEmail,
		SecretNickname:        state.Game.SecretNickname,
		SecretNation:          state.Game.SecretNation,
		Deadlines:             state.Game.Deadlines,
		ChatFlags:             state.Game.ChatFlags,
		AllocationMethod:      state.Game.AllocationMethod,
		NonCommitConsequences: state.Game.NonCommitConsequences,
		NMRConsequences:       state.Game.NMRConsequences,
		Ranking:               state.Game.Ranking,
	}

	if _, found := variants.Variants[game.Variant]; !found {
		return fmt.Errorf("Unknown variant for %+v", game)
	}

	if _, found := allocation.Methods[game.AllocationMethod]; !found {
		return fmt.Errorf("Unknown allocation method for %+v", game)
	}

	member := &Member{
		UserId:           kol.Id(c.Principal()),
		PreferredNations: state.Members[0].PreferredNations,
	}
	return c.Transact(func(c common.WSContext) error {
		if err := c.DB().Set(game); err != nil {
			return err
		}
		member.GameId = game.Id
		return c.DB().Set(member)
	})
}
