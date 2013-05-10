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

	router.Path("/login").Methods("GET").HandlerFunc(common.Login)
	router.Path("/logout").Methods("GET").HandlerFunc(common.Logout)

	router.Path("/user").MatcherFunc(wantsJSON).Methods("GET").HandlerFunc(common.GetUser)

	gamesRouter := router.PathPrefix("/games").MatcherFunc(wantsJSON).Subrouter()
	gamesRouter.Methods("GET").HandlerFunc(games.GetGames)

	http.Handle("/", router)
}
