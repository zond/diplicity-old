package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/user"
	"github.com/zond/diplicity/web"
	"github.com/zond/kcwraps/subs"
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
	router.HandleFunc("/token", server.Logger(server.Token))

	wsRouter := subs.NewRouter(server.DB())
	wsRouter.Resource("^/games/current$").Auth().
		Handle(subs.SubscribeType, game.SubscribeCurrent)
	wsRouter.Resource("^/games/open$").Auth().
		Handle(subs.SubscribeType, game.SubscribeOpen).
		Handle(subs.UpdateType, game.AddMember)
	wsRouter.Resource("^/user$").
		Handle(subs.SubscribeType, user.SubscribeEmail).
		Handle(subs.UpdateType, user.Update)
	wsRouter.Resource("^/games/(.*)/messages$").Auth()
	Handle(subs.SubscribeType, game.SubscribeMessages).
		Handle(subs.CreateType, game.CreateMessage)
	wsRouter.Resource("^/games/(.*)$").
		Handle(subs.SubscribeType, game.SubscribeGame).
		Handle(subs.DeleteType, game.DeleteMember)
	wsRouter.Resource("^/games$").Auth()
	Handle(subs.CreateType, game.Create)

	wsRouter.RPC("GetPossibleSources", game.GetPossibleSources).Auth()
	wsRouter.RPC("GetValidOrders", game.GetValidOrders).Auth()
	wsRouter.RPC("SetOrder", game.SetOrder).Auth()

	// The websocket
	router.Path("/ws").Handler(wsRouter)

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
