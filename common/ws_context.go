package common

import (
	"github.com/zond/kcwraps/subs"
	"github.com/zond/wsubs/gosubs"
)

type WSContext interface {
	subs.Context
	SendMail(from, subject, message string, recips ...string) error
}

type defaultWSContext struct {
	subs.Context
	web *Web
}

func (self *defaultWSContext) SendMail(from, subject, message string, recips ...string) error {
	return self.web.SendMail(from, subject, message, recips...)
}

type Router struct {
	*subs.Router
	web *Web
}

func NewRouter(web *Web) *Router {
	return &Router{
		Router: subs.NewRouter(web.DB()),
		web:    web,
	}
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
		return f(&defaultWSContext{
			Context: c,
			web:     self.web,
		})
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
			return f(&defaultWSContext{
				Context: c,
				web:     self.web,
			})
		}),
	}
}
