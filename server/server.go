package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/web"
	"net/http"
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

func main() {
	port := flag.Int("port", 8080, "The port to listen on")
	secret := flag.String("secret", defaultSecret, "The cookie store secret")
	env := flag.String("env", "development", "What environment to run")
	appcache := flag.Bool("appcache", true, "Whether to enable appcache")

	flag.Parse()

	if *env != "development" && *secret == defaultSecret {
		panic("Only development env can run with the default secret")
	}

	server := web.New().SetEnv(*env).SetSecret(*secret).SetAppcache(*appcache)

	router := mux.NewRouter()

	// Static content
	router.PathPrefix("/img").Handler(http.FileServer(http.Dir("")))
	router.HandleFunc("/js/{ver}/all", server.Logger(server.AllJs))
	router.HandleFunc("/css/{ver}/all", server.Logger(server.AllCss))
	router.HandleFunc("/diplicity.appcache", server.Logger(server.AppCache))

	// Login/logout
	router.HandleFunc("/login", server.Logger(server.Login))
	router.HandleFunc("/logout", server.Logger(server.Logout))
	router.HandleFunc("/openid", server.Logger(server.Openid))

	// The websocket
	router.Path("/ws").Handler(websocket.Handler(server.WS))

	// Everything else HTMLy
	router.MatcherFunc(wantsHTML).HandlerFunc(server.Logger(server.Index))

	addr := fmt.Sprintf("0.0.0.0:%v", *port)

	server.Infof("Listening to %v", addr)
	server.Fatalf("%v", http.ListenAndServe(addr, router))

}
