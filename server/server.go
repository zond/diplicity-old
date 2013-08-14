package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/web"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
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
		log.Printf("%v\t%v\t%v\t%v", self.request.Method, self.request.URL, self.status, time.Now().Sub(self.start))
	} else {
		log.Printf("%v\t%v\t%v\t%v\t%v", self.request.Method, self.request.URL, self.status, time.Now().Sub(self.start), err)
	}
}

func (self *loggingResponseWriter) inc() {
	log.Printf("%v\t%v", self.request.Method, self.request.URL)
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
	router := mux.NewRouter()

	// Static content
	router.PathPrefix("/img").Handler(http.FileServer(http.Dir("")))
	router.HandleFunc("/js/{ver}/all", logger(web.AllJs))
	router.HandleFunc("/css/{ver}/all", logger(web.AllCss))
	router.HandleFunc("/diplicity.appcache", logger(web.AppCache))

	// Login/logout
	router.HandleFunc("/login", logger(web.Login))
	router.HandleFunc("/logout", logger(web.Logout))
	router.HandleFunc("/openid", logger(web.Openid))

	// The websocket
	router.Path("/ws").Handler(websocket.Handler(web.WS))

	// Everything else HTMLy
	router.MatcherFunc(wantsHTML).HandlerFunc(logger(web.Index))

	var port int
	var err error
	if port, err = strconv.Atoi(web.PortEnv); err != nil {
		port = 80
	}
	addr := fmt.Sprintf("0.0.0.0:%v", port)

	log.Printf("Listening to %v", addr)
	log.Fatal(http.ListenAndServe(addr, router))

}
