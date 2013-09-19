package user

import (
	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/subs"
)

type User struct {
	Id    []byte
	Email string
}

func SubscribeEmail(c common.Context, s *subs.Subscription, email string) {
	s.Subscribe(&User{Id: []byte(email)})
}
