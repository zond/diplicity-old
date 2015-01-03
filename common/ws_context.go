package common

import (
	"github.com/zond/kcwraps/subs"
	"github.com/zond/wsubs/gosubs"
)

type WSContext interface {
	subs.SubContext
	BetweenTransactions(func(WSContext) error) error
	Transact(func(WSContext) error) error
	Mailer
	Env() string
	Diet() SkinnyContext
	Secret() string
}

func NewWSContext(c subs.Context, web *Web) WSContext {
	return &defaultWSContext{
		Context: c,
		web:     web,
	}
}

type defaultWSContext struct {
	subs.Context
	web *Web
}

func (self *defaultWSContext) Secret() string {
	return self.web.secret
}

func (self *defaultWSContext) Diet() SkinnyContext {
	return skinnyWSContext{WSContext: self}
}

func (self defaultWSContext) BetweenTransactions(f func(WSContext) error) (err error) {
	return self.Context.BetweenTransactions(func(c subs.Context) (err error) {
		self.Context = c
		return f(&self)
	})
}

func (self defaultWSContext) Transact(f func(c WSContext) error) error {
	return self.Context.Transact(func(c subs.Context) error {
		self.Context = c
		return f(&self)
	})
}

func (self *defaultWSContext) Env() string {
	return self.web.env
}

func (self *defaultWSContext) SendAddress() string {
	return self.web.smtpAccount
}

func (self *defaultWSContext) ReceiveAddress() string {
	return self.web.gmailAccount
}

func (self *defaultWSContext) SendMail(fromName, replyTo, subject, message string, recips []string) error {
	return self.web.SendMail(fromName, replyTo, subject, message, recips)
}

type Router struct {
	*subs.Router
	web *Web
}

func newRouter(web *Web) (result *Router) {
	result = &Router{
		Router: subs.NewRouter(web.DB()),
		web:    web,
	}
	return
}

type RPC struct {
	*gosubs.RPC
}

func (self *RPC) Auth() *RPC {
	self.RPC.Auth()
	return self
}

type Resource struct {
	*subs.Resource
	web *Web
}

func (self *Resource) Handle(op string, f func(c WSContext) error) *Resource {
	self.Resource.Handle(op, func(c subs.Context) error {
		return f(NewWSContext(c, self.web))
	})
	return self
}

func (self *Resource) Auth() *Resource {
	self.Resource.Auth()
	return self
}

func (self *Router) Resource(s string) *Resource {
	return &Resource{
		Resource: self.Router.Resource(s),
		web:      self.web,
	}
}

func (self *Router) RPC(m string, f func(c WSContext) (result interface{}, err error)) *RPC {
	return &RPC{
		RPC: self.Router.RPC(m, func(c subs.Context) (result interface{}, err error) {
			return f(NewWSContext(c, self.web))
		}),
	}
}
