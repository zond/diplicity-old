package srv

import (
	"fmt"

	"github.com/zond/unbolted"
	"github.com/zond/unbolted/pack"
)

type Context interface {
	pack.SubContext
	AfterTransaction(func(Context) error) error
	View(func(Context) error) error
	Update(func(Context) error) error
	Mailer
	Env() string
	Secret() string
	TX() *unbolted.TX
}

func NewContext(c pack.Context, web *Server) Context {
	return &defaultContext{
		Context: c,
		web:     web,
	}
}

type defaultContext struct {
	pack.Context
	web     *Server
	tx      *unbolted.TX
	writing bool
}

func (self *defaultContext) Secret() string {
	return self.web.secret
}

func (self defaultContext) AfterTransaction(f func(Context) error) (err error) {
	return self.Context.AfterTransaction(func(c pack.Context) (err error) {
		self.Context = c
		return f(&self)
	})
}

func (self *defaultContext) Env() string {
	return self.web.env
}

func (self *defaultContext) TX() *unbolted.TX {
	return self.tx
}

func (self defaultContext) Update(f func(c Context) error) error {
	if self.tx != nil {
		if !self.writing {
			return fmt.Errorf("%+v began with a View transaction!", self)
		}
		return f(&self)
	}
	return self.Context.Update(func(c pack.TXContext) error {
		self.writing = true
		self.tx = c.TX()
		return f(&self)
	})
}

func (self defaultContext) View(f func(c Context) error) error {
	if self.tx != nil {
		return f(&self)
	}
	return self.Context.View(func(c pack.TXContext) error {
		self.tx = c.TX()
		return f(&self)
	})
}

func (self *defaultContext) SendAddress() string {
	return self.web.smtpAccount
}

func (self *defaultContext) ReceiveAddress() string {
	return self.web.gmailAccount
}

func (self *defaultContext) SendMail(fromName, replyTo, subject, message string, recips []string) error {
	return self.web.SendMail(fromName, replyTo, subject, message, recips)
}
