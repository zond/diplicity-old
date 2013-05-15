package games

import (
	"appengine/datastore"
	"common"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func FetchOpenGames(w http.ResponseWriter, r *http.Request) {
	data := common.GetRequestData(w, r)
	if data.Authenticated() {
		common.SetContentType(w, "application/json; charset=UTF-8")
		common.MustEncodeJSON(w, GetOpenGamesForUser(data.Context, data.User.Email))
	}
}

func FetchGameMembers(w http.ResponseWriter, r *http.Request) {
	data := common.GetRequestData(w, r)
	if data.Authenticated() {
		common.SetContentType(w, "application/json; charset=UTF-8")
		common.MustEncodeJSON(w, GetGameMembersByUser(data.Context, data.User.Email))
	}
}

func DeleteGameMember(w http.ResponseWriter, r *http.Request) {
	data := common.GetRequestData(w, r)
	if data.Authenticated() {
		// validate that the key is ok
		member_id, err := datastore.DecodeKey(mux.Vars(r)["member_id"])
		if err != nil {
			panic(err)
		}
		GetGameMemberById(data.Context, member_id).ValidatedDelete(data.Context, data.User.Email)
		common.SetContentType(w, "application/json; charset=UTF-8")
		w.WriteHeader(204)
	}
}

func UpdateGameMemberWithGame(w http.ResponseWriter, r *http.Request) {
	data := common.GetRequestData(w, r)
	if data.Authenticated() {
		// validate that the key is ok
		member_id, err := datastore.DecodeKey(mux.Vars(r)["member_id"])
		if err != nil {
			panic(err)
		}
		// validate that the member exists
		current := GetGameMemberById(data.Context, member_id)
		if current == nil {
			panic(fmt.Errorf("No GameMember with id %v found", member_id))
		}
		// validate that user owns game
		if current.Game.Id.Parent().StringID() != data.User.Email {
			panic(fmt.Errorf("%+v not owned by %v", current.Game, data.User.Email))
		}
		// validate that user is member
		if current.Email != data.User.Email {
			panic(fmt.Errorf("%+v is not %v", current, data.User.Email))
		}
		var upload GameMember
		common.MustDecodeJSON(r.Body, &upload)
		// validate that user owns upload
		if upload.Game.Id.Parent().StringID() != data.User.Email {
			panic(fmt.Errorf("%+v not owned by %v", upload.Game, data.User.Email))
		}
		// validate that user is upload
		if current.Email != data.User.Email {
			panic(fmt.Errorf("%+v is not %+v", upload, current))
		}
		// validate that upload is game
		if !current.Game.Id.Equal(upload.Game.Id) {
			panic(fmt.Errorf("%+v is not %+v", upload.Game, current.Game))
		}
		common.SetContentType(w, "application/json; charset=UTF-8")
		if _, err := current.CopyFrom(&upload).SaveWithGame(data.Context, data.User.Email); err != nil {
			data.Response.WriteHeader(500)
			data.Context.Infof("%v", err)
			common.MustEncodeJSON(w, err)
		} else {
			common.MustEncodeJSON(w, current)
		}
	}
}

func CreateGameMemberWithGame(w http.ResponseWriter, r *http.Request) {
	data := common.GetRequestData(w, r)
	if data.Authenticated() {
		var member GameMember
		common.MustDecodeJSON(r.Body, &member)
		member.Email = data.User.Email
		common.SetContentType(w, "application/json; charset=UTF-8")
		if _, err := (&member).SaveWithGame(data.Context, data.User.Email); err != nil {
			data.Response.WriteHeader(500)
			data.Context.Infof("%v", err)
			common.MustEncodeJSON(w, err)
		} else {
			common.MustEncodeJSON(w, member)
		}
	}
}
