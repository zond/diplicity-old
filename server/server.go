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
	secret := flag.String("secret", gosubs.Secret, "The cookie store secret")
	env := flag.String("env", "development", "What environment to run")
	appcache := flag.Bool("appcache", true, "Whether to enable appcache")

	flag.Parse()

	if *env != "development" && *secret == gosubs.Secret {
		panic("Only development env can run with the default secret")
	}

	server := web.New().SetEnv(*env).SetAppcache(*appcache)

	router := mux.NewRouter()

	// Login/logout
	server.Handle(router.Path("/login"), server.Login)
	server.Handle(router.Path("/logout"), server.Logout)
	server.Handle(router.Path("/openid"), server.Openid)
	server.Handle(router.Path("/token"), server.Token)

	// Resource routes for the WebSocket
	wsRouter := subs.NewRouter(server.DB())
	if *env == "development" {
		wsRouter.SetDevMode()
	}
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

	// RPC routes for the WebSocket
	wsRouter.RPC("GetPossibleSources", game.GetPossibleSources).Auth()
	wsRouter.RPC("GetValidOrders", game.GetValidOrders).Auth()
	wsRouter.RPC("SetOrder", game.SetOrder).Auth()
	wsRouter.RPC("Commit", game.CommitPhase).Auth()
	wsRouter.RPC("Uncommit", game.UncommitPhase).Auth()

	// The websocket
	router.Path("/ws").Handler(wsRouter)

	// Static content
	server.Handle(router.Path("/js/{ver}/all"), server.AllJs)
	server.Handle(router.Path("/css/{ver}/all"), server.AllCss)
	server.Handle(router.Path("/diplicity.appcache"), server.AppCache)
	server.HandleStatic(router, "static")

	// Admin
	server.AdminHandle(router.Path("/admin/games/{game_id}"), server.AdminGetGame)

	// Everything else HTMLy
	server.Handle(router.MatcherFunc(wantsHTML), server.Index)

	addr := fmt.Sprintf("0.0.0.0:%v", *port)

	server.Infof("Listening to %v  (env=%v, appcache=%v)", addr, *env, *appcache)
	if err := server.Start(); err != nil {
		panic(err)
	}
	server.Fatalf("%v", http.ListenAndServe(addr, router))

}
