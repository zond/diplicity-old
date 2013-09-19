package game

import (
	"encoding/base64"
	"fmt"
	"github.com/zond/diplicity/common"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
	"net/url"
)

type Minutes int

type Game struct {
	Id    []byte
	Phase []byte

	Closed  bool `kol:"index"`
	Started bool `kol:"index"`
	Ended   bool `kol:"index"`
	Variant string
	EndYear int
	Private bool `kol:"index"`

	Deadlines map[dip.PhaseType]Minutes

	ChatFlags map[dip.PhaseType]common.ChatFlag
}

func DeleteMember(c common.Context, gameId, email string) {
	if err := c.DB().Transact(func(d *kol.DB) error {
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
		member := Member{}
		if _, err := d.Query().Where(kol.And{kol.Equals{"Game", base64DecodedId}, kol.Equals{"User", []byte(email)}}).First(&member); err != nil {
			return err
		}
		if !game.Started {
			if err := d.Del(&member); err != nil {
				return err
			}
			left := []Member{}
			if err := d.Query().Where(kol.Equals{"Game", game.Id}).All(&left); err != nil {
				return err
			}
			if len(left) == 0 {
				if err := d.Del(&game); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		c.Errorf("Unable to delete member: %v", err)
	}
}

func AddMember(c common.Context, gameId, email string) {
	if err := c.DB().Transact(func(d *kol.DB) error {
		base64DecodedId, err := base64.StdEncoding.DecodeString(gameId)
		if err != nil {
			return err
		}
		game := Game{Id: base64DecodedId}
		if err := d.Get(&game); err != nil {
			return err
		}
		variant, found := common.VariantMap[game.Variant]
		if !found {
			return fmt.Errorf("Unknown variant %v", game.Variant)
		}
		already := []Member{}
		if err := d.Query().Where(kol.Equals{"Game", base64DecodedId}).All(&already); err != nil {
			return err
		}
		if len(already) < len(variant.Nations) {
			member := Member{Game: base64DecodedId, User: []byte(email)}
			if err := d.Set(&member); err != nil {
				return err
			}
			if len(already) == len(variant.Nations)-1 {
				game.Started = true
				game.Closed = true
				if err := d.Set(&game); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		c.Errorf("Unable to add member: %v", err)
	}
}

func Create(c common.Context, j common.JSON, creator string) {
	var state gameState
	j.Overwrite(&state)

	game := &Game{
		Variant:   state.Game.Variant,
		EndYear:   state.Game.EndYear,
		Private:   state.Game.Private,
		Deadlines: state.Game.Deadlines,
		ChatFlags: state.Game.ChatFlags,
	}

	member := &Member{
		User: []byte(creator),
	}
	c.DB().Transact(func(d *kol.DB) error {
		if err := d.Set(game); err != nil {
			return err
		}
		member.Game = game.Id
		return d.Set(member)
	})
}

func (self *Game) Updated(d *kol.DB, old *Game) {
	members := []Member{}
	if err := d.Query().Where(kol.Equals{"Game", self.Id}).All(&members); err != nil {
		panic(err)
	}
	for _, member := range members {
		d.EmitUpdate(&member)
	}
}

type Phase struct {
	Id   []byte
	Game []byte `kol:"index"`

	Season  dip.Season
	Year    int
	Type    dip.PhaseType
	Ordinal int
}

func (self *Phase) Updated(d *kol.DB, old *Phase) {
	g := Game{Id: self.Game}
	if err := d.Get(&g); err != nil {
		panic(err)
	}
	d.EmitUpdate(&g)
}

type Member struct {
	Id   []byte
	User []byte `kol:"index"`
	Game []byte `kol:"index"`

	Nation dip.Nation
}

type gameState struct {
	*Game
	Member *Member
	Phase  *Phase
}

func SubscribeCurrent(c common.Context, s *subs.Subscription, email string) {
	s.Query = s.DB().Query().Where(kol.Equals{"User", []byte(email)})
	s.Call = func(i interface{}, op string) {
		members := i.([]*Member)
		states := []gameState{}
		for _, member := range members {
			if op == common.DeleteType {
				states = append(states, gameState{
					Member: member,
				})
			} else {
				game := &Game{Id: member.Game}
				if err := s.DB().Get(game); err != nil {
					panic(err)
				}
				if !game.Ended {
					phase := &Phase{Id: game.Phase}
					if err := s.DB().Get(phase); err != nil {
						if err == kol.NotFound {
							phase = nil
						} else {
							panic(err)
						}
					}
					states = append(states, gameState{
						Game:   game,
						Member: member,
						Phase:  phase,
					})
				}
			}
		}
		s.Send(states, op)
	}
	s.Subscribe(new(Member))
}

func SubscribeOpen(c common.Context, s *subs.Subscription, email string) {
	s.Query = s.DB().Query().Where(kol.And{kol.Equals{"Closed", false}, kol.Equals{"Private", false}})
	s.Call = func(i interface{}, op string) {
		var members []Member
		games := i.([]*Game)
		states := []gameState{}
		for _, game := range games {
			members = nil
			s.DB().Query().Where(kol.And{kol.Equals{"User", []byte(email)}, kol.Equals{"Game", game.Id}}).All(&members)
			if len(members) == 0 {
				phase := &Phase{Id: game.Phase}
				if err := s.DB().Get(phase); err != nil {
					if err == kol.NotFound {
						phase = nil
					} else {
						panic(err)
					}
				}
				states = append(states, gameState{
					Game:  game,
					Phase: phase,
				})
			}
		}
		s.Send(states, op)
	}
	s.Subscribe(new(Game))
}
