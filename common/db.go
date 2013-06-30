package common

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/zond/kcwraps/kol"
	"reflect"
)

var DB *kol.DB

const (
	fetchType = "Fetch"
)

func init() {
	var err error
	if DB, err = kol.New("diplicity"); err != nil {
		panic(err)
	}
}

func Unsubscribe(ws *websocket.Conn, url string) {
	DB.Unsubscribe(fmt.Sprintf("%v/%v", ws.Request().RemoteAddr, url))
}

func subscriber(ws *websocket.Conn, url string) (s func(i interface{}, op string)) {
	return func(i interface{}, op string) {
		if err := websocket.JSON.Send(ws, JsonMessage{
			Type: op,
			Object: &ObjectMessage{
				Data: i,
				URL:  url,
			},
		}); err != nil {
			Unsubscribe(ws, url)
		}
	}
}

func Subscribe(ws *websocket.Conn, url string, obj interface{}) {
	s := subscriber(ws, url)
	if err := DB.Subscribe(fmt.Sprintf("%v/%v", ws.Request().RemoteAddr, url), obj, kol.AllOps, func(i interface{}, op kol.Operation) {
		s(i, op.String())
	}); err != nil {
		panic(err)
	}
	if err := DB.Get(obj); err != nil {
		if err != kol.NotFound {
			panic(err)
		}
	} else {
		s(obj, fetchType)
	}
}

func SubscribeQuery(ws *websocket.Conn, url string, q *kol.Query, obj interface{}) {
	s := subscriber(ws, url)
	if err := q.Subscribe(fmt.Sprintf("%v/%v", ws.Request().RemoteAddr, url), obj, kol.AllOps, func(i interface{}, op kol.Operation) {
		s([]interface{}{i}, op.String())
	}); err != nil {
		panic(err)
	}
	slice := reflect.MakeSlice(reflect.TypeOf(obj), 0, 0).Interface()
	if err := q.All(slice); err != nil {
		panic(err)
	} else {
		s(slice, fetchType)
	}
}
