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
)

func wantsJSON(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "application/json"
}

func wantsHTML(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "text/html"
}

func main() {
	router := mux.NewRouter()

	// Static content
	router.PathPrefix("/img").Handler(http.FileServer(http.Dir("")))
	router.HandleFunc("/js/{ver}/all", web.AllJs)
	router.HandleFunc("/css/{ver}/all", web.AllCss)
	router.HandleFunc("/diplicity.appcache", web.AppCache)
	router.HandleFunc("/reload", web.Reload)

	// Login/logout
	router.HandleFunc("/login", web.Login)
	router.HandleFunc("/openid", web.Openid)

	// The websocket
	router.Path("/ws").Handler(websocket.Handler(web.WS))

	// Everything else HTMLy
	router.MatcherFunc(wantsHTML).HandlerFunc(web.Index)

	var port int
	var err error
	if port, err = strconv.Atoi(web.PortEnv); err != nil {
		port = 80
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", port), router))

}
