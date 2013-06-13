package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/zond/diplicity/common"
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
	fmt.Println("got", ws)
}

func main() {
	router := mux.NewRouter()

	// Static content
	router.HandleFunc("/js/{ver}/all", allJs)
	router.HandleFunc("/css/{ver}/all", allCss)
	router.HandleFunc("/diplicity.appcache", appCache)
	router.HandleFunc("/reload", reload)

	// The websocket
	router.Path("/ws").Handler(websocket.Handler(wsHandler))

	// Everything else HTMLy
	router.MatcherFunc(wantsHTML).HandlerFunc(index)

	var port int
	var err error
	if port, err = strconv.Atoi(os.Getenv("PORT")); err != nil {
		port = 80
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", port), router))

}
