package user

import (
	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/subs"
)

type User struct {
	Id    []byte
	Email string
}

func EmailSubscription(s *subs.Subscription, email string) *common.Subscription {
	return &common.Subscription{
		Name:       s.Name(),
		Subscriber: s.Call,
		Query:      nil,
		Object:     &User{Id: []byte(email)},
	}
}
