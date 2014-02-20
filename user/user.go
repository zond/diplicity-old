package user

import (
	"fmt"
	"time"
	"code.google.com/p/go.net/websocket"

	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
)

type User struct {
	Id              kol.Id
	Email           string
	Nickname        string
	MissedDeadlines int
	HeldDeadlines   int
	Ranking         float64

	CreatedAt time.Time
	UpdatedAt time.Time
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

func SubscribeEmail(c subs.Context) error {
	if c.Principal() == "" {
		return websocket.JSON.Send(c.Conn(), subs.Message{
			Type: subs.FetchType,
			Object: &subs.Object{
				URI:  c.Match()[0],
				Data: &User{},
			},
		})
	}
	s := c.Pack().New(c.Match()[0])
	return s.Subscribe(&User{Id: kol.Id(c.Principal())})
}

func Update(c subs.Context) (err error) {
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
	err = c.DB().Set(current)
	return
}

func EnsureUser(c subs.Context) (result *User, err error) {
	result = &User{Id: kol.Id(c.Principal())}
	if err = c.DB().Get(result); err == kol.NotFound {
		result.Email = c.Principal()
		err = c.DB().Set(result)
	}
	return
}
