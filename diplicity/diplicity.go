package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zond/diplicity/epoch"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/srv"
	"github.com/zond/diplicity/user"
	"github.com/zond/wsubs"
	"github.com/zond/ziprot"
)

func wantsJSON(r *http.Request, m *mux.RouteMatch) bool {
	return srv.MostAccepted(r, "text/html", "Accept") == "application/json"
}

func wantsHTML(r *http.Request, m *mux.RouteMatch) bool {
	mostAccepted := srv.MostAccepted(r, "text/html", "Accept")
	return mostAccepted == "text/html" || mostAccepted == "*/*"
}

func main() {
	port := flag.Int("port", 8080, "The port to listen on")
	secret := flag.String("secret", srv.DefaultSecret, "The cookie store secret")
	gmailAccount := flag.String("gmail_account", "", "The GMail account to use for sending and receiving message email")
	gmailPassword := flag.String("gmail_password", "", "The GMail account password")
	env := flag.String("env", srv.Development, "What environment to run")
	db := flag.String("db", "diplicity.db", "The path to the database file to use")
	logOutput := flag.String("log", "-", "Where to send the log output")
	smtpAccount := flag.String("smtp_account", "", "What From-address to put in the outgoing email")
	smtpHost := flag.String("smtp_host", "", "What host to use when sending out email")
	schedule := flag.Bool("schedule", true, "Schedule unresolved phases at startup")
	oauthClientSecret := flag.String("oauth_client_secret", "", "The client secret of your OAuth credentials in Google Cloud. See https://developers.google.com/accounts/docs/OpenIDConnect")
	oauthClientId := flag.String("oauth_client_id", "", "The client id of your OAuth credentials in Google Cloud. See See https://developers.google.com/accounts/docs/OpenIDConnect")

	flag.Parse()

	if *logOutput != "-" {
		z, err := ziprot.New(*logOutput)
		if err != nil {
			panic(err)
		}
		log.SetOutput(z.MaxFiles(10).MaxSize(1024 * 1024 * 256))
	}

	server, err := srv.NewServer(*secret, *env, *db)
	if err != nil {
		panic(err)
	}
	server.SetGMail(*gmailAccount, *gmailPassword, game.IncomingMail).SetSMTP(*smtpHost, *smtpAccount)

	if *env != srv.Development {
		if *oauthClientSecret == "" {
			server.Errorf("No oauth_client_secret provided, you will not be able to use Google single sign on")
		}
		if *oauthClientId == "" {
			server.Errorf("No oauth_client_id provided, you will not be able to use Google single sign on")
		}
	}

	router := mux.NewRouter()

	// Login/logout
	server.Handle(router.Path("/login"), user.Login(*oauthClientId))
	server.Handle(router.Path("/logout"), user.Logout)
	server.Handle(router.Path("/oauth2callback"), user.OAuth2Callback(*oauthClientId, *oauthClientSecret))
	server.Handle(router.Path("/token"), user.Token)

	// Resource routes for the WebSocket
	wsRouter := server.Router()
	wsRouter.Resource("^/games/mine$").
		Handle(wsubs.SubscribeType, game.SubscribeMine)
	wsRouter.Resource("^/games/open$").
		Handle(wsubs.SubscribeType, game.SubscribeOthersOpen)
	wsRouter.Resource("^/games/closed$").
		Handle(wsubs.SubscribeType, game.SubscribeOthersClosed)
	wsRouter.Resource("^/games/finished$").
		Handle(wsubs.SubscribeType, game.SubscribeOthersFinished)
	wsRouter.Resource("^/user$").
		Handle(wsubs.SubscribeType, user.SubscribeEmail).
		Handle(wsubs.UpdateType, user.Update).Auth()
	wsRouter.Resource("^/games/(.+)/(\\d+)$").
		Handle(wsubs.SubscribeType, game.SubscribeGamePhase)
	wsRouter.Resource("^/games/(.+)/messages$").
		Handle(wsubs.SubscribeType, game.SubscribeMessages).
		Handle(wsubs.CreateType, game.CreateMessage).Auth()
	wsRouter.Resource("^/games/(.+)$").
		Handle(wsubs.SubscribeType, game.SubscribeGame).
		Handle(wsubs.DeleteType, game.DeleteMember).Auth().
		Handle(wsubs.UpdateType, game.AddMember).Auth()
	wsRouter.Resource("^/games$").
		Handle(wsubs.CreateType, game.Create).Auth()

	// RPC routes for the WebSocket
	wsRouter.RPC("SetOrder", game.SetOrder).Auth()
	wsRouter.RPC("Commit", game.CommitPhase).Auth()
	wsRouter.RPC("Uncommit", game.UncommitPhase).Auth()
	wsRouter.RPC("See", game.SeeMessage).Auth()

	// The websocket
	router.Path("/ws").Handler(wsRouter)

	// Static content
	if err := server.HandleStatic(router, "bower_components"); err != nil {
		panic(err)
	}
	if err := server.HandleStatic(router, "components"); err != nil {
		panic(err)
	}
	if err := server.HandleStatic(router, "static"); err != nil {
		panic(err)
	}

	// js exports from go
	server.Handle(router.Path("/go.js").Methods("GET"), server.Go2JS)

	// Admin
	server.AdminHandle(router.Path("/admin/games/{game_id}/rollback/{until}").Methods("POST"), game.AdminRollback)
	server.AdminHandle(router.Path("/admin/games/{game_id}").Methods("GET"), game.AdminGetGame)
	server.AdminHandle(router.Path("/admin/games/{game_id}/nations/{nation}/options").Methods("GET"), game.AdminGetOptions)
	server.AdminHandle(router.Path("/admin/users").Methods("POST"), user.AdminCreateUser)
	server.AdminHandle(router.Path("/admin/games/{game_id}/recalc").Methods("POST"), game.AdminRecalcOptions)
	server.AdminHandle(router.Path("/admin/games/reindex").Methods("POST"), game.AdminReindexGames)
	server.AdminHandle(router.Path("/admin/users/setrank1").Methods("POST"), user.AdminSetRank1)
	server.DevHandle(router.Path("/admin/become").Methods("POST"), user.DevBecome)
	server.DevHandle(router.Path("/admin/login").Methods("GET"), user.DevLogin)

	server.Handle(router.Path("/resolve/{variant}").Methods("POST"), game.Resolve)

	// Unsubscribe
	server.Handle(router.Path("/unsubscribe/{unsubscribe_tag}").Methods("GET"), game.UnsubscribeEmails)

	// Everything else HTMLy
	server.Handle(router.MatcherFunc(wantsHTML), server.Index)

	addr := fmt.Sprintf("0.0.0.0:%v", *port)

	if err := server.Start(); err != nil {
		panic(err)
	}
	if err := epoch.Start(server.Context()); err != nil {
		panic(err)
	}
	if *schedule {
		if err := game.ScheduleUnresolvedPhases(server.Context()); err != nil {
			panic(err)
		}
	}
	server.Infof("Listening to %v (env=%#v, gmail_account=%#v, smtp_account=%#v, smtp_host=%#v)", addr, *env, *gmailAccount, *smtpAccount, *smtpHost)
	server.Fatalf("%v", http.ListenAndServe(addr, router))

}
