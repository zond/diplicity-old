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
	router.HandleFunc("/js/{ver}/all", allJs)
	router.HandleFunc("/css/{ver}/all", allCss)
	router.HandleFunc("/diplicity.appcache", appCache)
	router.HandleFunc("/", index)

	router.Path("/user").MatcherFunc(wantsJSON).Methods("GET").HandlerFunc(common.GetUser)

	gamesRouter := router.PathPrefix("/games").MatcherFunc(wantsJSON).Subrouter()
	gamesRouter.Methods("GET").HandlerFunc(games.GetGames)

	http.Handle("/", router)
}
