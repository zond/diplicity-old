package user

import (
	"github.com/zond/diplicity/db"
)

type User struct {
	Id    []byte
	Email string
}

func SubscribeEmail(s *db.Subscription, email interface{}) {
	if email == nil {
		s.Call(&User{}, db.FetchType)
	} else {
		db.Subscribe(s.Name(), s.Call, &User{Id: []byte(email.(string))})
		s.Register()
	}
}
