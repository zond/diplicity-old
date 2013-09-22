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

func wantsJSON(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "application/json"
}

func wantsHTML(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "text/html"
}

func main() {
	port := flag.Int("port", 8080, "The port to listen on")
	secret := flag.String("secret", web.DefaultSecret, "The cookie store secret")
	env := flag.String("env", "development", "What environment to run")
	appcache := flag.Bool("appcache", true, "Whether to enable appcache")

	flag.Parse()

	if *env != "development" && *secret == web.DefaultSecret {
		panic("Only development env can run with the default secret")
	}

	server := web.New().SetEnv(*env).SetSecret(*secret).SetAppcache(*appcache)

	router := mux.NewRouter()

	// Login/logout
	router.HandleFunc("/login", server.Logger(server.Login))
	router.HandleFunc("/logout", server.Logger(server.Logout))
	router.HandleFunc("/openid", server.Logger(server.Openid))

	// The websocket
	router.Path("/ws").Handler(websocket.Handler(server.WS))

	// Static content
	router.HandleFunc("/js/{ver}/all", server.Logger(server.AllJs))
	router.HandleFunc("/css/{ver}/all", server.Logger(server.AllCss))
	router.HandleFunc("/diplicity.appcache", server.Logger(server.AppCache))
	server.HandleStatic(router, "static")

	// Everything else HTMLy
	router.MatcherFunc(wantsHTML).HandlerFunc(server.Logger(server.Index))

	addr := fmt.Sprintf("0.0.0.0:%v", *port)

	server.Infof("Listening to %v  (env=%v, appcache=%v)", addr, *env, *appcache)
	server.Fatalf("%v", http.ListenAndServe(addr, router))

}
