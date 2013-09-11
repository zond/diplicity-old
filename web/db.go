package web

import (
	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/kol"
	"reflect"
)

const (
	FetchType = "Fetch"
)

func (self *Web) Subscribe(s *common.Subscription) {
	if s != nil {
		if err := self.db.Subscribe(s.Name, s.Object, kol.AllOps, func(i interface{}, op kol.Operation) {
			s.Subscriber(i, op.String())
		}); err != nil {
			panic(err)
		}
		if err := self.db.Get(s.Object); err != nil {
			if err != kol.NotFound {
				panic(err)
			}
		} else {
			s.Subscriber(s.Object, FetchType)
		}
	}
}

func (self *Web) SubscribeQuery(s *common.Subscription) {
	if s != nil {
		if err := s.Query.Subscribe(s.Name, s.Object, kol.AllOps, func(i interface{}, op kol.Operation) {
			slice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(s.Object)), 1, 1)
			slice.Index(0).Set(reflect.ValueOf(i))
			s.Subscriber(slice.Interface(), op.String())
		}); err != nil {
			panic(err)
		}
		slice := reflect.New(reflect.SliceOf(reflect.TypeOf(s.Object))).Interface()
		if err := s.Query.All(slice); err != nil {
			panic(err)
		} else {
			s.Subscriber(reflect.ValueOf(slice).Elem().Interface(), FetchType)
		}
	}
}
