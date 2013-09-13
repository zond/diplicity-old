package user

import (
	"github.com/zond/kcwraps/subs"
)

type User struct {
	Id    []byte
	Email string
}

func SubscribeEmail(s *subs.Subscription, email string) {
	s.Subscribe(&User{Id: []byte(email)})
}
