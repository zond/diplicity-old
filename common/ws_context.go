package common

import (
	"fmt"

	"github.com/zond/diplicity/translation"
	"github.com/zond/kcwraps/subs"
	"github.com/zond/wsubs/gosubs"
)

type Mailer interface {
	SendMail(from, subject, message string, recips ...string) error
	MailAddress() string
}

type WSContext interface {
	subs.SubContext
	BetweenTransactions(func(c WSContext))
	Transact(func(c WSContext) error) error
	Mailer
	I(phrase string, args ...interface{}) (string, error)
	Env() string
	Diet() SkinnyContext
}

func NewWSContext(c subs.Context, web *Web) WSContext {
	return &defaultWSContext{
		Context:      c,
		web:          web,
		translations: translation.GetTranslations(GetLanguage(c.Conn().Request())),
	}
}

type defaultWSContext struct {
	subs.Context
	web          *Web
	translations map[string]string
}

func (self *defaultWSContext) Diet() SkinnyContext {
	return skinnyWSContext{WSContext: self}
}

func (self defaultWSContext) BetweenTransactions(f func(c WSContext)) {
	self.Context.BetweenTransactions(func(c subs.Context) {
		self.Context = c
		f(&self)
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

func (self *defaultWSContext) I(phrase string, args ...interface{}) (result string, err error) {
	pattern, ok := self.translations[phrase]
	if !ok {
		err = fmt.Errorf("Found no translation for %v", phrase)
		result = err.Error()
		return
	}
	if len(args) > 0 {
		result = fmt.Sprintf(pattern, args...)
		return
	}
	result = pattern
	return
}

func (self *defaultWSContext) MailAddress() string {
	return self.web.gmailAccount
}

func (self *defaultWSContext) SendMail(from, subject, message string, recips ...string) error {
	return self.web.SendMail(from, subject, message, recips...)
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
	result.Router.LogLevel = web.logLevel
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
