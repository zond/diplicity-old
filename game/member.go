package game

import (
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
)

type Member struct {
	Id     kol.Id
	UserId kol.Id `kol:"index"`
	GameId kol.Id `kol:"index"`

	Nation           dip.Nation
	PreferredNations []dip.Nation
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

type Members []Member

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

func (self Members) toStates(c common.Context, g *Game, email string) (result []MemberState) {
	result = make([]MemberState, len(self))
	for index, member := range self {
		cpy := member
		if string(cpy.UserId) != email {
			cpy.UserId = nil
			cpy.PreferredNations = nil
		}
		if string(cpy.UserId) != email && g.SecretNation {
			cpy.Nation = ""
		}
		result[index] = MemberState{
			Member: &cpy,
			User:   &user.User{},
		}
		if !g.SecretEmail || !g.SecretNickname {
			foundUser := user.EnsureUser(c, string(member.UserId))
			if !g.SecretEmail {
				result[index].User.Email = foundUser.Email
				result[index].User.Id = foundUser.Id
			}
			if !g.SecretNickname {
				result[index].User.Nickname = foundUser.Nickname
			}
		}
	}
	return
}
