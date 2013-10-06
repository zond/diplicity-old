package game

import (
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
)

type Member struct {
	Id     []byte
	UserId []byte `kol:"index"`
	GameId []byte `kol:"index"`

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
