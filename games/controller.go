package games

import (
	"common"
	"net/http"
)

func GetGames(w http.ResponseWriter, r *http.Request) {
	data := common.GetRequestData(w, r)
	if data.Authenticated() {
		common.SetContentType(w, "application/json; charset=UTF-8")
		common.MustEncodeJSON(w, GetGamesByUser(data.Context, data.User.Email))
	}
}
