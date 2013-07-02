package db

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/kol"
	"reflect"
	"sync"
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

type Subscriber func(i interface{}, op string)

// SubscriptionPack keeps track of the subscriptions for a websocket
type SubscriptionPack struct {
	lock *sync.Mutex
	subs map[string]*Subscription
	ws   *websocket.Conn
}

func NewSubscriptionPack(ws *websocket.Conn) *SubscriptionPack {
	return &SubscriptionPack{
		lock: new(sync.Mutex),
		subs: make(map[string]*Subscription),
		ws:   ws,
	}
}

func (self *SubscriptionPack) subscriptionName(url string) string {
	return fmt.Sprintf("%v/%v", self.ws.Request().RemoteAddr, url)
}

// NewSubscription creates a new subscription for a pack
func (self *SubscriptionPack) NewSubscription(url string) (result *Subscription) {
	result = &Subscription{
		pack: self,
		url:  url,
	}
	return
}

// Unsubscribe unsubscribes the given url
func (self SubscriptionPack) Unsubscribe(url string) {
	if s, found := self.subs[self.subscriptionName(url)]; found {
		s.unsubscribe()
	}
}

func (self SubscriptionPack) UnsubscribeAll() {
	self.lock.Lock()
	defer self.lock.Unlock()
	for _, s := range self.subs {
		DB.Unsubscribe(s.Name())
	}
	self.subs = make(map[string]*Subscription)
}

// Subscription is a single subscription for a websocket
type Subscription struct {
	pack *SubscriptionPack
	url  string
}

// Name is the identifier of the subscription in the kol DB
func (self Subscription) Name() string {
	return self.pack.subscriptionName(self.url)
}

// Call does whatever the subscription wants to do when it gets an event
func (self Subscription) Call(i interface{}, op string) {
	if err := websocket.JSON.Send(self.pack.ws, common.JsonMessage{
		Type: op,
		Object: &common.ObjectMessage{
			Data: i,
			URL:  self.url,
		},
	}); err != nil {
		self.unsubscribe()
	}
}

func (self *Subscription) Register() {
	self.pack.lock.Lock()
	defer self.pack.lock.Unlock()
	self.pack.subs[self.Name()] = self
}

func (self *Subscription) unsubscribe() {
	self.pack.lock.Lock()
	defer self.pack.lock.Unlock()
	delete(self.pack.subs, self.Name())
	DB.Unsubscribe(self.Name())
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
	slice := reflect.New(reflect.SliceOf(reflect.TypeOf(obj))).Interface()
	if err := q.All(slice); err != nil {
		panic(err)
	} else {
		s(reflect.ValueOf(slice).Elem().Interface(), FetchType)
	}
}
