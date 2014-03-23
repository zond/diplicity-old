package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/epoch"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/user"
	"github.com/zond/wsubs/gosubs"
	"github.com/zond/ziprot"
)

func wantsJSON(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "application/json"
}

func wantsHTML(r *http.Request, m *mux.RouteMatch) bool {
	mostAccepted := common.MostAccepted(r, "text/html", "Accept")
	return mostAccepted == "text/html" || mostAccepted == "*/*"
}

func main() {
	port := flag.Int("port", 8080, "The port to listen on")
	secret := flag.String("secret", common.DefaultSecret, "The cookie store secret")
	gmailAccount := flag.String("gmail_account", "", "The GMail account to use for sending and receiving message email")
	gmailPassword := flag.String("gmail_password", "", "The GMail account password")
	env := flag.String("env", common.Development, "What environment to run")
	db := flag.String("db", "diplicity", "The path to the database file to use")
	appcache := flag.Bool("appcache", true, "Whether to enable appcache")
	logOutput := flag.String("log", "-", "Where to send the log output")

	flag.Parse()

	if *logOutput != "-" {
		z, err := ziprot.New(*logOutput)
		if err != nil {
			panic(err)
		}
		log.SetOutput(z.MaxFiles(10).MaxSize(1024 * 1024 * 256))
	}

	server, err := common.NewWeb(*secret, *env, *db)
	if err != nil {
		panic(err)
	}
	server.SetAppcache(*appcache).SetGMail(*gmailAccount, *gmailPassword, game.IncomingMail)

	router := mux.NewRouter()

	// Login/logout
	server.Handle(router.Path("/login"), user.Login)
	server.Handle(router.Path("/logout"), user.Logout)
	server.Handle(router.Path("/openid"), user.Openid)
	server.Handle(router.Path("/token"), user.Token)

	// Resource routes for the WebSocket
	wsRouter := server.Router()
	wsRouter.Resource("^/games/current$").
		Handle(gosubs.SubscribeType, game.SubscribeCurrent).Auth()
	wsRouter.Resource("^/games/open$").
		Handle(gosubs.SubscribeType, game.SubscribeOpen).Auth()
	wsRouter.Resource("^/user$").
		Handle(gosubs.SubscribeType, user.SubscribeEmail).
		Handle(gosubs.UpdateType, user.Update).Auth()
	wsRouter.Resource("^/games/(.*)/messages$").
		Handle(gosubs.SubscribeType, game.SubscribeMessages).
		Handle(gosubs.CreateType, game.CreateMessage).Auth()
	wsRouter.Resource("^/games/(.*)$").
		Handle(gosubs.SubscribeType, game.SubscribeGame).
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
	wsRouter.RPC("See", game.SeeMessage).Auth()

	// The websocket
	router.Path("/ws").Handler(wsRouter)

	// Static content
	server.Handle(router.Path("/js/{ver}/all"), server.AllJs)
	server.Handle(router.Path("/css/{ver}/all"), server.AllCss)
	server.Handle(router.Path("/diplicity.appcache"), server.AppCache)
	if err := server.HandleStatic(router, "static"); err != nil {
		panic(err)
	}

	// Admin
	server.AdminHandle(router.Path("/admin/games/{game_id}").Methods("GET"), game.AdminGetGame)
	server.AdminHandle(router.Path("/admin/users").Methods("POST"), user.AdminCreateUser)

	// Unsubscribe
	server.Handle(router.Path("/unsubscribe/{unsubscribe_tag}").Methods("GET"), game.UnsubscribeEmails)

	// Everything else HTMLy
	server.Handle(router.MatcherFunc(wantsHTML), server.Index)

	addr := fmt.Sprintf("0.0.0.0:%v", *port)

	if err := server.Start(); err != nil {
		panic(err)
	}
	if err := epoch.Start(server.Diet()); err != nil {
		panic(err)
	}
	if err := game.ScheduleUnresolvedPhases(server.Diet()); err != nil {
		panic(err)
	}
	server.Infof("Listening to %v  (env=%#v, appcache=%#v, gmail_account=%#v)", addr, *env, *appcache, *gmailAccount)
	server.Fatalf("%v", http.ListenAndServe(addr, router))

}
