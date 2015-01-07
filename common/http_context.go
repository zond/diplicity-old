package common

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/zond/unbolted"
)

type HTTPContext struct {
	response http.ResponseWriter
	request  *http.Request
	session  *sessions.Session
	vars     map[string]string
	web      *Server
}

func (self *HTTPContext) Env() string {
	return self.web.Env()
}

func (self *HTTPContext) DB() *unbolted.DB {
	return self.web.db
}

func (self *HTTPContext) Secret() string {
	return self.web.secret
}

func (self *HTTPContext) Session() *sessions.Session {
	return self.session
}

func (self *HTTPContext) SetContentType(t string) {
	self.Resp().Header().Set("Content-Type", t)
	self.Resp().Header().Set("Vary", "Accept")
	self.Resp().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	self.Resp().Header().Set("Pragma", "no-cache")
	self.Resp().Header().Set("Expires", "0")
}

func (self *HTTPContext) Fatalf(format string, args ...interface{}) {
	self.web.Fatalf(format, args...)
}

func (self *HTTPContext) Errorf(format string, args ...interface{}) {
	self.web.Errorf(format, args...)
}

func (self *HTTPContext) Infof(format string, args ...interface{}) {
	self.web.Infof(format, args...)
}

func (self *HTTPContext) Debugf(format string, args ...interface{}) {
	self.web.Debugf(format, args...)
}

func (self *HTTPContext) Tracef(format string, args ...interface{}) {
	self.web.Tracef(format, args...)
}

func (self *HTTPContext) RenderJSON(i interface{}) (err error) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return
	}
	self.SetContentType("application/json; charset=UTF-8")
	_, err = self.Resp().Write(b)
	return
}

func (self *HTTPContext) Vars() map[string]string {
	return self.vars
}

func (self *HTTPContext) Resp() http.ResponseWriter {
	return self.response
}

func (self *HTTPContext) Req() *http.Request {
	return self.request
}

func (self *HTTPContext) Close() {
	self.session.Save(self.request, self.response)
}
