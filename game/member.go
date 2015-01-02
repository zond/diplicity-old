package game

import (
	"fmt"
	"strings"
	"time"

	"github.com/zond/diplicity/game/meta"
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

	Options interface{}

	Committed bool
	NoOrders  bool
	NoWait    bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Members []Member

func (self Members) Len() int {
	return len(self)
}

func (self Members) Less(i, j int) bool {
	return self[j].CreatedAt.Before(self[i].CreatedAt)
}

func (self Members) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self Members) Get(email string) *Member {
	for index, _ := range self {
		if string(self[index].UserId) == email {
			return &self[index]
		}
	}
	return nil
}

func (self Members) Contains(email string) bool {
	for _, member := range self {
		if string(member.UserId) == email {
			return true
		}
	}
	return false
}

func (self Members) ToStates(d *kol.DB, g *Game, email string, isAdmin bool) (result []MemberState, err error) {
	result = make([]MemberState, len(self))
	isMember := false
	for _, member := range self {
		if member.UserId.Equals(kol.Id(email)) {
			isMember = true
			break
		}
	}
	for index, member := range self {
		var state *MemberState
		if state, err = member.ToState(d, g, email, isMember, isAdmin); err != nil {
			return
		}
		result[index] = *state
	}
	return
}

func (self *Member) ToState(d *kol.DB, g *Game, email string, isMember bool, isAdmin bool) (result *MemberState, err error) {
	result = &MemberState{
		Member: &Member{
			Id: self.Id,
		},
		User: &user.User{},
	}
	privacyConfig := PrivacyConfig{}
	switch g.State {
	case meta.GameStateCreated:
		privacyConfig = g.PrivacyConfigs[BeforePhaseType]
	case meta.GameStateStarted:
		privacyConfig = g.PrivacyConfigs[DuringPhaseType]
	case meta.GameStateEnded:
		privacyConfig = g.PrivacyConfigs[AfterPhaseType]
	default:
		panic(fmt.Errorf("Unknown game state for %+v", g))
	}
	isMe := string(self.UserId) == email
	if isAdmin || isMe || !privacyConfig.SecretNation {
		result.Member.Nation = self.Nation
	}
	if isAdmin || isMe || !privacyConfig.SecretEmail || !privacyConfig.SecretNickname {
		foundUser := &user.User{Id: self.UserId}
		if err = d.Get(foundUser); err != nil {
			return
		}
		if isAdmin || (isMember && (isMe || !privacyConfig.SecretEmail)) {
			result.User.Email = foundUser.Email
		}
		if isAdmin || (isMe || !privacyConfig.SecretNickname) {
			result.User.Nickname = foundUser.Nickname
		}
		if isAdmin || isMe {
			result.Member.Committed = self.Committed
			result.Member.Options = self.Options
			result.Member.NoOrders = self.NoOrders
		}
	}
	return
}

func (self *Member) ShortName(game *Game, user *user.User) string {
	if game.State == meta.GameStateCreated {
		if user.Nickname != "" && !game.PrivacyConfigs[BeforePhaseType].SecretNickname {
			return user.Nickname
		}
		if user.Email != "" && !game.PrivacyConfigs[BeforePhaseType].SecretEmail {
			return strings.Split(user.Email, "@")[0]
		}
		return "Anonymous"
	}
	return string(self.Nation)
}

func (self *Member) Deleted(d *kol.DB) {
	g := Game{Id: self.GameId}
	if err := d.Get(&g); err == nil {
		d.EmitUpdate(&g)
	} else if err != kol.NotFound {
		panic(err)
	}
}

func (self *Member) Updated(d *kol.DB, old *Member) {
	if old != self {
		g := Game{Id: self.GameId}
		if err := d.Get(&g); err != nil {
			panic(err)
		}
		d.EmitUpdate(&g)
	}
}

func (self *Member) Created(d *kol.DB) {
	g := Game{Id: self.GameId}
	if err := d.Get(&g); err != nil {
		panic(err)
	}
	d.EmitUpdate(&g)
}

func (self *Member) ReliabilityDelta(d *kol.DB, i int) (err error) {
	user := &user.User{Id: self.UserId}
	if err = d.Get(user); err != nil {
		return
	}
	if i > 0 {
		user.HeldDeadlines += i
	} else {
		user.MissedDeadlines -= i
	}
	if err = d.Set(user); err != nil {
		return
	}
	return
}

func (self Members) Disallows(d *kol.DB, asking *user.User) (result bool, err error) {
	if asking == nil {
		return
	}
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
