package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/web"
	"log"
	"net/http"
	"os"
	"strconv"
)

func wantsJSON(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "application/json"
}

func wantsHTML(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "text/html"
}

func wsHandler(ws *websocket.Conn) {
	var message string
	for err := websocket.Message.Receive(ws, &message); err == nil; err = websocket.Message.Receive(ws, &message) {
		fmt.Println(message)
	}
}

func main() {
	router := mux.NewRouter()

	// Static content
	router.PathPrefix("/img").Handler(http.FileServer(http.Dir("")))
	router.HandleFunc("/js/{ver}/all", web.AllJs)
	router.HandleFunc("/css/{ver}/all", web.AllCss)
	router.HandleFunc("/diplicity.appcache", web.AppCache)
	router.HandleFunc("/reload", web.Reload)

	// The websocket
	router.Path("/ws").Handler(websocket.Handler(wsHandler))

	// Everything else HTMLy
	router.MatcherFunc(wantsHTML).HandlerFunc(web.Index)

	var port int
	var err error
	if port, err = strconv.Atoi(os.Getenv("PORT")); err != nil {
		port = 80
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", port), router))

}
