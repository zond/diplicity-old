package web

import (
	"common"
	"games"
	"github.com/gorilla/mux"
	"net/http"
)

func wantsJSON(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "application/json"
}

func wantsHTML(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "text/html"
}

func init() {
	router := mux.NewRouter()

	// Pre flights
	router.Methods("OPTIONS").HandlerFunc(preflight)

	// Static content
	router.HandleFunc("/js/{ver}/all", allJs)
	router.HandleFunc("/css/{ver}/all", allCss)
	router.HandleFunc("/diplicity.appcache", appCache)
	router.HandleFunc("/reload", reload)
	router.HandleFunc("/", index)

	// Login/logout redirects
	router.Path("/login").Methods("GET").HandlerFunc(common.Login)
	router.Path("/logout").Methods("GET").HandlerFunc(common.Logout)

	/*
		JSON endpoints
	*/

	// Logged in user
	router.Path("/user").MatcherFunc(wantsJSON).Methods("GET").HandlerFunc(common.GetUser)

	// Games
	gamesRouter := router.PathPrefix("/games").MatcherFunc(wantsJSON).Subrouter()

	// Games of which the user is a member
	membersRouter := gamesRouter.PathPrefix("/member").Subrouter()
	membersRouter.Methods("GET").HandlerFunc(games.GetGameMembers)
	membersRouter.Methods("POST").HandlerFunc(games.CreateGameMemberWithGame)

	// Forming games for the user
	formingRouter := gamesRouter.PathPrefix("/open").Subrouter()
	formingRouter.Methods("GET").HandlerFunc(games.GetOpenGames)

	http.Handle("/", router)
}
