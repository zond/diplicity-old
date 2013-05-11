package games

import (
	"common"
	"net/http"
)

func GetGames(w http.ResponseWriter, r *http.Request) {
	data := common.GetRequestData(w, r)
	if data.Authenticated() {
		common.SetContentType(w, "application/json; charset=UTF-8")
		common.MustEncodeJSON(w, GetGameMembersByUser(data.Context, data.User.Email))
	}
}

func CreateGame(w http.ResponseWriter, r *http.Request) {
	data := common.GetRequestData(w, r)
	if data.Authenticated() {
		var member GameMember
		common.MustDecodeJSON(r.Body, &member)
		common.SetContentType(w, "application/json; charset=UTF-8")
		if _, err := (&member).CreateWithGame(data.Context, data.User.Email); err != nil {
			data.Response.WriteHeader(500)
			data.Context.Infof("%v", err)
			common.MustEncodeJSON(w, err)
		} else {
			common.MustEncodeJSON(w, member)
		}
	}
}
