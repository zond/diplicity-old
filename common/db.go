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

const (
	FetchType = "Fetch"
)

func subscriptionName(ws *websocket.Conn, url string) string {
	return fmt.Sprintf("%v/%v", ws.Request().RemoteAddr, url)
}

func Unsubscribe(ws *websocket.Conn, url string) {
	DB.Unsubscribe(subscriptionName(ws, url))
}

type Subscriber func(i interface{}, op string)

type WSSubscription struct {
	ws  *websocket.Conn
	url string
}

func NewWSSubscription(ws *websocket.Conn, url string) WSSubscription {
	return WSSubscription{
		ws:  ws,
		url: url,
	}
}

func (self WSSubscription) Name() string {
	return subscriptionName(self.ws, self.url)
}

func (self WSSubscription) Call(i interface{}, op string) {
	if err := websocket.JSON.Send(self.ws, JsonMessage{
		Type: op,
		Object: &ObjectMessage{
			Data: i,
			URL:  self.url,
		},
	}); err != nil {
		self.unsubscribe()
	}
}

func (self WSSubscription) unsubscribe() {
	Unsubscribe(self.ws, self.url)
}

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
	slice := reflect.MakeSlice(reflect.TypeOf(obj), 0, 0).Interface()
	if err := q.All(slice); err != nil {
		panic(err)
	} else {
		s(slice, FetchType)
	}
}
