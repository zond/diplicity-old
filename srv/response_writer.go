package srv

import (
	"compress/gzip"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
	request    *http.Request
	start      time.Time
	status     int
	web        *Server
}

func (self *responseWriter) email() string {
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

func (self *responseWriter) WriteHeader(i int) {
	self.status = i
	self.ResponseWriter.WriteHeader(i)
}

func (self *responseWriter) Write(b []byte) (n int, err error) {
	if self.gzipWriter != nil {
		return self.gzipWriter.Write(b)
	}
	return self.ResponseWriter.Write(b)
}

func (self *responseWriter) Close() (err error) {
	return self.gzipWriter.Close()
}

func (self *responseWriter) log(err interface{}) {
	if err == nil {
		self.web.Infof("%v\t%v\t%v\t%v\t%v\t%v", self.request.Method, self.request.URL, self.request.RemoteAddr, self.email(), self.status, time.Now().Sub(self.start))
	} else {
		self.web.Errorf("%v\t%v\t%v\t%v\t%v\t%v\t%v", self.request.Method, self.request.URL, self.request.RemoteAddr, self.email(), self.status, time.Now().Sub(self.start), err)
	}
}

func (self *responseWriter) inc() {
	self.web.Infof("%v\t%v\t%v\t%v", self.request.Method, self.request.URL, self.request.RemoteAddr, self.email())
}
