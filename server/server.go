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
	"github.com/zond/wsubs/gosubs"
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
	wsRouter.LogLevel = gosubs.DebugLevel
	wsRouter.Resource("^/games/current$").
		Handle(gosubs.SubscribeType, game.SubscribeCurrent).Auth()
	wsRouter.Resource("^/games/open$").
		Handle(gosubs.SubscribeType, game.SubscribeOpen).Auth()
	wsRouter.Resource("^/user$").
		Handle(gosubs.SubscribeType, user.SubscribeEmail).
		Handle(gosubs.UpdateType, user.Update).Auth()
	wsRouter.Resource("^/games/(.*)/messages$").
		Handle(gosubs.SubscribeType, game.SubscribeMessages).Auth().
		Handle(gosubs.CreateType, game.CreateMessage).Auth()
	wsRouter.Resource("^/games/(.*)$").
		Handle(gosubs.SubscribeType, game.SubscribeGame).Auth().
		Handle(gosubs.DeleteType, game.DeleteMember).Auth().
		Handle(gosubs.UpdateType, game.AddMember).Auth()
	wsRouter.Resource("^/games$").
		Handle(gosubs.CreateType, game.Create).Auth()

	wsRouter.RPC("GetPossibleSources", game.GetPossibleSources).Auth()
	wsRouter.RPC("GetValidOrders", game.GetValidOrders).Auth()
	wsRouter.RPC("SetOrder", game.SetOrder).Auth()
	wsRouter.RPC("Commit", game.CommitPhase).Auth()
	wsRouter.RPC("Uncommit", game.UncommitPhase).Auth()

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
