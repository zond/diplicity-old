package user

import (
	"github.com/zond/diplicity/common"
)

type User struct {
	Id    []byte
	Email string
}

func SubscribeEmail(s common.WSSubscription, email interface{}) {
	if email == nil {
		s.Call(&User{}, common.FetchType)
	} else {
		common.Subscribe(s.Name(), s.Call, &User{Id: []byte(email.(string))})
	}
}
