package user

import (
	"github.com/zond/diplicity/db"
	"github.com/zond/kcwraps/subs"
)

type User struct {
	Id    []byte
	Email string
}

func SubscribeEmail(s *subs.Subscription, email interface{}) {
	if email == nil {
		s.Call(&User{}, db.FetchType)
	} else {
		db.Subscribe(s.Name(), s.Call, &User{Id: []byte(email.(string))})
	}
}
