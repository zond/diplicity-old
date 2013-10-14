package user

import (
	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
)

type User struct {
	Id              kol.Id
	Email           string
	Nickname        string
	MissedDeadlines int
	HeldDeadlines   int
	Ranking         int
}

type Blacklisting struct {
	Id          kol.Id
	Blacklister kol.Id
	Blacklistee kol.Id
}

func SubscribeEmail(c common.Context, s *subs.Subscription, email string) error {
	return s.Subscribe(&User{Id: []byte(email)})
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
	user := &User{Id: []byte(email)}
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
