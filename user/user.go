package user

import (
	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
	"time"
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

func SubscribeEmail(c common.Context, s *subs.Subscription, email string) error {
	return s.Subscribe(&User{Id: kol.Id(email)})
}

func Update(c common.Context, j subs.JSON, email string) (err error) {
	var user User
	j.Overwrite(&user)
	current := &User{Id: user.Id}
	if err = c.DB().Get(current); err != nil {
		return
	}
	current.Nickname = user.Nickname
	err = c.DB().Set(current)
	return
}

func EnsureUser(c common.Context, email string) *User {
	user := &User{Id: kol.Id(email)}
	if err := c.DB().Get(user); err == kol.NotFound {
		user.Email = email
		if err = c.DB().Set(user); err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}
	return user
}
