package web

import (
	"net/http"
	"sync/atomic"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	request *http.Request
	start   time.Time
	status  int
	web     *Web
}

func (self *loggingResponseWriter) email() string {
	session, err := self.web.sessionStore.Get(self.request, SessionName)
	if err != nil {
		self.web.Errorf("\t%v\t%v\t%v", self.request.URL, self.request.RemoteAddr, err)
	}
	email := ""
	emailIf, loggedIn := session.Values[SessionEmail]
	if loggedIn {
		email = emailIf.(string)
	}
	return email
}

func (self *loggingResponseWriter) WriteHeader(i int) {
	self.status = i
	self.ResponseWriter.WriteHeader(i)
}

func (self *loggingResponseWriter) log(err interface{}) {
	if err == nil {
		self.web.Infof("%v\t%v\t%v\t%v\t%v\t%v", self.request.Method, self.request.URL, self.request.RemoteAddr, self.email(), self.status, time.Now().Sub(self.start))
	} else {
		self.web.Errorf("%v\t%v\t%v\t%v\t%v\t%v\t%v", self.request.Method, self.request.URL, self.request.RemoteAddr, self.email(), self.status, time.Now().Sub(self.start), err)
	}
}

func (self *loggingResponseWriter) inc() {
	self.web.Infof("%v\t%v\t%v\t%v", self.request.Method, self.request.URL, self.request.RemoteAddr, self.email())
}

func (self *Web) Logger(f func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			request:        r,
			start:          time.Now(),
			status:         200,
			web:            self,
		}
		var i int64
		defer func() {
			atomic.StoreInt64(&i, 1)
			lw.log(recover())
		}()
		go func() {
			time.Sleep(time.Second)
			if atomic.CompareAndSwapInt64(&i, 0, 1) {
				lw.inc()
			}
		}()
		f(lw, r)
	}
}
