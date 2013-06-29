package common

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
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

func Unsubscribe(ws *websocket.Conn, url string) {
	DB.Unsubscribe(fmt.Sprintf("%v/%v", ws.RemoteAddr().String(), url))
}

func Subscribe(ws *websocket.Conn, url string, q *kol.Query, obj interface{}) {
	s := func(i interface{}, op kol.Operation) {
		if err := websocket.JSON.Send(ws, JsonMessage{
			Type:   op.String(),
			Object: i,
		}); err != nil {
			Unsubscribe(ws, url)
		}
	}
	if err := q.Subscribe(fmt.Sprintf("%v/%v", ws.RemoteAddr().String(), url), obj, kol.AllOps, func(i interface{}, op kol.Operation) {
		s([]interface{}{i}, op)
	}); err != nil {
		panic(err)
	}
	slice := reflect.MakeSlice(reflect.TypeOf(obj), 0, 0).Interface()
	if err := q.All(slice); err != nil {
		s(slice, kol.Create)
	}
}
