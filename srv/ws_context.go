package srv

import (
	"github.com/zond/unbolted"
	"github.com/zond/unbolted/pack"
	"github.com/zond/wsubs"
)

func NewWSContext(c pack.Context, web *Server) WSContext {
	return &defaultWSContext{
		Context: c,
		web:     web,
	}
}

type defaultWSContext struct {
	pack.Context
	web *Server
}

func (self *defaultWSContext) Secret() string {
	return self.web.secret
}

func (self *defaultWSContext) Diet() SkinnyContext {
	return &skinnyWSContext{WSContext: self}
}

func (self defaultWSContext) AfterTransaction(f func(WSContext) error) (err error) {
	return self.Context.AfterTransaction(func(c pack.Context) (err error) {
		self.Context = c
		return f(&self)
	})
}

func (self *defaultWSContext) Env() string {
	return self.web.env
}

type defaultWSTXContext struct {
	*defaultWSContext
	tx *unbolted.TX
}

func (self *defaultWSTXContext) TXDiet() SkinnyTXContext {
	return &skinnyTXContext{
		SkinnyContext: self.Diet(),
		tx:            self.tx,
	}
}

func (self *defaultWSTXContext) TX() *unbolted.TX {
	return self.tx
}

func (self *defaultWSContext) Update(f func(c WSTXContext) error) error {
	return self.Context.Update(func(c pack.TXContext) error {
		return f(&defaultWSTXContext{
			defaultWSContext: self,
			tx:               c.TX(),
		})
	})
}

func (self *defaultWSContext) View(f func(c WSTXContext) error) error {
	return self.Context.View(func(c pack.TXContext) error {
		return f(&defaultWSTXContext{
			defaultWSContext: self,
			tx:               c.TX(),
		})
	})
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
	*pack.Router
	web *Server
}

func newRouter(web *Server) (result *Router) {
	result = &Router{
		Router: pack.NewRouter(web.DB()),
		web:    web,
	}
	return
}

type RPC struct {
	*wsubs.RPC
}

func (self *RPC) Auth() *RPC {
	self.RPC.Auth()
	return self
}

type Resource struct {
	*pack.Resource
	web *Server
}

func (self *Resource) Handle(op string, f func(c WSContext) error) *Resource {
	self.Resource.Handle(op, func(c pack.Context) error {
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
		RPC: self.Router.RPC(m, func(c pack.Context) (result interface{}, err error) {
			return f(NewWSContext(c, self.web))
		}),
	}
}

type skinnyWSContext struct {
	WSContext
}

func (self *skinnyWSContext) AfterTransaction(f func(SkinnyContext) error) (err error) {
	return self.WSContext.AfterTransaction(func(c WSContext) error {
		self.WSContext = c
		return f(self)
	})
}

func (self *skinnyWSContext) Update(f func(c SkinnyTXContext) error) error {
	return self.WSContext.Update(func(c WSTXContext) error {
		return f(&skinnyTXContext{
			SkinnyContext: self,
			tx:            c.TX(),
		})
	})
}

func (self *skinnyWSContext) View(f func(c SkinnyTXContext) error) error {
	return self.WSContext.View(func(c WSTXContext) error {
		return f(&skinnyTXContext{
			SkinnyContext: self,
			tx:            c.TX(),
		})
	})
}
