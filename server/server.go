package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/web"
	"github.com/zond/kcwraps/kol"
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

type messageType string

const (
	subscribeType   messageType = "subscribe"
	unsubscribeType messageType = "unsubscribe"
	createdType     messageType = "created"
	updatedType     messageType = "updated"
	deletedType     messageType = "deleted"
)

type subscribeMessage struct {
	URI string
}

type jsonMessage struct {
	Type      messageType
	Subscribe *subscribeMessage
	Object    interface{}
}

func wsHandler(ws *websocket.Conn) {
	var message jsonMessage
	for err := websocket.JSON.Receive(ws, &message); err == nil; err = websocket.JSON.Receive(ws, &message) {
		switch message.Type {
		case subscribeType:
			switch message.Subscribe.URI {
			case "/games/open":
				game.SubscribeOpen(ws.LocalAddr().String(), func(g *game.Game, op kol.Operation) {
					var typ messageType
					switch op {
					case kol.Create:
						typ = createdType
					case kol.Update:
						typ = updatedType
					case kol.Delete:
						typ = deletedType
					}
					if err := websocket.JSON.Send(ws, jsonMessage{
						Type:   typ,
						Object: g,
					}); err != nil {
						game.UnsubscribeOpen(ws.LocalAddr().String())
					}
				})
			default:
				fmt.Printf("Unrecognized URI: %+v\n", message)
			}
		case unsubscribeType:
		default:
			fmt.Printf("Unrecognized message: %+v\n", message)
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
