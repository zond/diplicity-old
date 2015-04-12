package srv

import (
	"github.com/zond/unbolted/pack"
	"github.com/zond/wsubs"
)

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

func (self *Resource) Handle(op string, f func(c Context) error) *Resource {
	self.Resource.Handle(op, func(c pack.Context) error {
		return f(NewContext(c, self.web))
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

func (self *Router) RPC(m string, f func(c Context) (result interface{}, err error)) *RPC {
	return &RPC{
		RPC: self.Router.RPC(m, func(c pack.Context) (result interface{}, err error) {
			return f(NewContext(c, self.web))
		}),
	}
}
