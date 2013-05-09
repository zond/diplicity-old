package games

import (
	"common"
	"net/http"
)

func GetGames(w http.ResponseWriter, r *http.Request) {
	common.SetContentType(w, "application/json; charset=UTF-8")
	common.MustEncodeJSON(w, []string{})
}
