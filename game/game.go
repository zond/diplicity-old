package game

import (
	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/kol"
)

type Minutes int

type Game struct {
	Id      []byte
	Closed  bool `kol:"index"`
	Started bool `kol:"index"`
	Variant string
	EndYear int
	Private bool `kol:"index"`
	Owner   []byte

	Deadlines map[string]Minutes

	ChatFlags map[string]common.ChatFlag
}

func UnsubscribeOpen(name string) {
	common.DB.Unsubscribe(name)
}

func SubscribeOpen(name string, f func(g *Game, op kol.Operation)) {
	var g Game
	if err := common.DB.Query().Filter(kol.Equals{"Closed", false}).Subscribe(name, &g, kol.AllOps, func(i interface{}, op kol.Operation) {
		f(i.(*Game), op)
	}); err != nil {
		panic(err)
	}
}
