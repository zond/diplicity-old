package games

import (
	"appengine"
	"appengine/datastore"
)

const (
	game        = "Game"
	allGamesKey = "Games{All}"
)

type Games []Game

type Game struct {
	Id      *datastore.Key `json:"id" datastore:"-"`
	Variant string         `json:"variant"`
}

func GetGamesByUser(c appengine.Context, email string) (result Games) {
	return make(Games, 0)
}
