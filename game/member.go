package game

import (
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
	"time"
)

type Member struct {
	Id     kol.Id
	UserId kol.Id `kol:"index"`
	GameId kol.Id `kol:"index"`

	Nation           dip.Nation
	PreferredNations []dip.Nation

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Members []Member

func (self Members) toStates(c common.Context, g *Game, email string) (result []MemberState) {
	result = make([]MemberState, len(self))
	for index, member := range self {
		state := member.toState(c, g, email)
		result[index] = *state
	}
	return
}

func (self *Member) toState(c common.Context, g *Game, email string) (result *MemberState) {
	result = &MemberState{
		Member: &Member{},
		User:   &user.User{},
	}
	secretNation := false
	secretEmail := false
	secretNickname := false
	var flag common.SecretFlag
	switch g.State {
	case common.GameStateCreated:
		flag = common.SecretBeforeGame
	case common.GameStateStarted:
		flag = common.SecretDuringGame
	case common.GameStateEnded:
		flag = common.SecretAfterGame
	default:
		panic(fmt.Errorf("Unknown game state for %+v", g))
	}
	secretNation, secretEmail, secretNickname = g.SecretNation&flag == flag, g.SecretEmail&flag == flag, g.SecretNickname&flag == flag
	me := string(self.UserId) == email
	if me || !secretNation {
		result.Member.Nation = self.Nation
	}
	if me || !secretEmail || !secretNickname {
		foundUser := user.EnsureUser(c.DB(), string(self.UserId))
		if me || !secretEmail {
			result.User.Email = foundUser.Email
		}
		if me || !secretNickname {
			result.User.Nickname = foundUser.Nickname
		}
	}
	return
}

func (self *Member) Deleted(d *kol.DB) {
	g := Game{Id: self.GameId}
	if err := d.Get(&g); err == nil {
		d.EmitUpdate(&g)
	} else if err != kol.NotFound {
		panic(err)
	}
}

func (self *Member) Created(d *kol.DB) {
	g := Game{Id: self.GameId}
	if err := d.Get(&g); err != nil {
		panic(err)
	}
	d.EmitUpdate(&g)
}

func (self Members) Disallows(d *kol.DB, asking *user.User) (result bool, err error) {
	var askerList map[string]bool
	if askerList, err = asking.Blacklistings(d); err != nil {
		return
	}
	for _, member := range self {
		if askerList[member.UserId.String()] {
			result = true
			return
		}
	}
	for _, member := range self {
		memberUser := &user.User{Id: member.UserId}
		if err = d.Get(memberUser); err != nil {
			return
		}
		var memberList map[string]bool
		if memberList, err = memberUser.Blacklistings(d); err != nil {
			return
		}
		if memberList[asking.Id.String()] {
			result = true
			return
		}
	}
	return
}
