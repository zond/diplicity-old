package user

import (
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

func Update(d *kol.DB) (err error) {
	var user User
	c.Data().Overwrite(&user)
	current := &User{Id: user.Id}
	if err = d.Get(current); err != nil {
		return
	}
	current.Nickname = user.Nickname
	err = d.Set(current)
	return
}

func EnsureUser(d *kol.DB, email string) *User {
	user := &User{Id: kol.Id(email)}
	if err := d.Get(user); err == kol.NotFound {
		user.Email = email
		if err = d.Set(user); err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}
	return user
}
