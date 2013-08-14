package db

import (
	"github.com/zond/kcwraps/kol"
	"reflect"
)

var DB *kol.DB

func init() {
	var err error
	if DB, err = kol.New("diplicity"); err != nil {
		panic(err)
	}
}

type Subscriber func(i interface{}, op string)

const (
	FetchType = "Fetch"
)

func Subscribe(name string, s Subscriber, obj interface{}) {
	if err := DB.Subscribe(name, obj, kol.AllOps, func(i interface{}, op kol.Operation) {
		s(i, op.String())
	}); err != nil {
		panic(err)
	}
	if err := DB.Get(obj); err != nil {
		if err != kol.NotFound {
			panic(err)
		}
	} else {
		s(obj, FetchType)
	}
}

func SubscribeQuery(name string, s Subscriber, q *kol.Query, obj interface{}) {
	if err := q.Subscribe(name, obj, kol.AllOps, func(i interface{}, op kol.Operation) {
		s([]interface{}{i}, op.String())
	}); err != nil {
		panic(err)
	}
	slice := reflect.New(reflect.SliceOf(reflect.TypeOf(obj))).Interface()
	if err := q.All(slice); err != nil {
		panic(err)
	} else {
		s(reflect.ValueOf(slice).Elem().Interface(), FetchType)
	}
}
