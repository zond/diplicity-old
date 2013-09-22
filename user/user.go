package user

import (
	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
)

type User struct {
	Id       []byte
	Email    string
	Nickname string
}

func SubscribeEmail(c common.Context, s *subs.Subscription, email string) {
	s.Subscribe(&User{Id: []byte(email)})
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
