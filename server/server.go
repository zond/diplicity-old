package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/web"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	defaultSecret = "something very secret"
)

func wantsJSON(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "application/json"
}

func wantsHTML(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "text/html"
}

type loggingResponseWriter struct {
	http.ResponseWriter
	request *http.Request
	start   time.Time
	status  int
}

func (self *loggingResponseWriter) WriteHeader(i int) {
	self.status = i
	self.ResponseWriter.WriteHeader(i)
}

func (self *loggingResponseWriter) log(err interface{}) {
	if err == nil {
		common.Infof("%v\t%v\t%v\t%v", self.request.Method, self.request.URL, self.status, time.Now().Sub(self.start))
	} else {
		common.Errorf("%v\t%v\t%v\t%v\t%v", self.request.Method, self.request.URL, self.status, time.Now().Sub(self.start), err)
	}
}

func (self *loggingResponseWriter) inc() {
	common.Infof("%v\t%v", self.request.Method, self.request.URL)
}

func logger(f func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			request:        r,
			start:          time.Now(),
			status:         200,
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

func main() {
	port := flag.Int("port", 8080, "The port to listen on")
	secret := flag.String("secret", defaultSecret, "The cookie store secret")
	env := flag.String("env", "development", "What environment to run")

	flag.Parse()

	if *env != "development" && *secret == defaultSecret {
		panic("Only development env can run with the default secret")
	}

	if *env == "development" {
		common.LogLevel = 100
	} else {
		common.LogLevel = 0
	}

	server := web.New(*env, *secret)

	router := mux.NewRouter()

	// Static content
	router.PathPrefix("/img").Handler(http.FileServer(http.Dir("")))
	router.HandleFunc("/js/{ver}/all", logger(server.AllJs))
	router.HandleFunc("/css/{ver}/all", logger(server.AllCss))
	router.HandleFunc("/diplicity.appcache", logger(server.AppCache))

	// Login/logout
	router.HandleFunc("/login", logger(server.Login))
	router.HandleFunc("/logout", logger(server.Logout))
	router.HandleFunc("/openid", logger(server.Openid))

	// The websocket
	router.Path("/ws").Handler(websocket.Handler(server.WS))

	// Everything else HTMLy
	router.MatcherFunc(wantsHTML).HandlerFunc(logger(server.Index))

	addr := fmt.Sprintf("0.0.0.0:%v", *port)

	common.Infof("Listening to %v", addr)
	common.Fatalf("%v", http.ListenAndServe(addr, router))

}
