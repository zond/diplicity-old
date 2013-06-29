package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/web"
	"io"
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
	log.Printf("%v connected", ws.RemoteAddr())
	var message common.JsonMessage
	var err error
	for {
		if err = websocket.JSON.Receive(ws, &message); err == nil {
			switch message.Type {
			case common.SubscribeType:
				switch message.Subscribe.URI {
				case "/games/open":
					common.Subscribe(ws, message.Subscribe.URI, game.Open(), new(game.Game))
				default:
					log.Printf("Unrecognized URI: %+v", message.Subscribe.URI)
				}
			case common.UnsubscribeType:
				common.Unsubscribe(ws, message.Subscribe.URI)
			default:
				log.Printf("Unrecognized message Type: %+v", message.Type)
			}
		} else if err == io.EOF {
			log.Printf("%v disconnected", ws.RemoteAddr())
			break
		} else {
			log.Println(err)
		}
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

	// Login/logout
	router.HandleFunc("/login", web.Login)

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
