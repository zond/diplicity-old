package user

import (
	"fmt"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/wsubs/gosubs"
)

type Users []User

type User struct {
	Id                   kol.Id
	Email                string
	Nickname             string
	MessageEmailDisabled bool
	PhaseEmailDisabled   bool
	MissedDeadlines      int
	HeldDeadlines        int
	Ranking              float64
	DiplicityHost        string

	LastLoginAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (self *User) Reliability() float64 {
	return float64(self.HeldDeadlines+1) / float64(self.MissedDeadlines+1)
}

func (self *User) Blacklistings(d *kol.DB) (result map[string]bool, err error) {
	result = map[string]bool{}
	var blacklistings []Blacklisting
	if err = d.Query().Where(kol.Equals{"Blacklister", self.Id}).All(&blacklistings); err != nil {
		return
	}
	for _, blacklisting := range blacklistings {
		result[blacklisting.Blacklistee.String()] = true
	}
	return
}

type Blacklisting struct {
	Id          kol.Id
	Blacklister kol.Id
	Blacklistee kol.Id
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func SubscribeEmail(c common.WSContext) error {
	if c.Principal() == "" {
		return websocket.JSON.Send(c.Conn(), gosubs.Message{
			Type: gosubs.FetchType,
			Object: &gosubs.Object{
				URI:  c.Match()[0],
				Data: &User{},
			},
		})
	}
	s := c.Pack().New(c.Match()[0])
	return s.Subscribe(&User{Id: kol.Id(c.Principal())})
}

func Update(c common.WSContext) (err error) {
	var user User
	c.Data().Overwrite(&user)
	current := &User{Id: user.Id}
	if err = c.DB().Get(current); err != nil {
		return
	}
	if current.Email != c.Principal() {
		err = fmt.Errorf("Unauthorized")
		return
	}
	current.Nickname = user.Nickname
	current.MessageEmailDisabled = user.MessageEmailDisabled
	current.PhaseEmailDisabled = user.PhaseEmailDisabled
	err = c.DB().Set(current)
	return
}
