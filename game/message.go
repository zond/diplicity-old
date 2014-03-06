package game

import (
	"time"

	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
)

type Messages []Message

func (self Messages) Len() int {
	return len(self)
}

func (self Messages) Less(j, i int) bool {
	return self[i].CreatedAt.Before(self[j].CreatedAt)
}

func (self Messages) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type Message struct {
	Id         kol.Id
	GameId     kol.Id `kol:"index"`
	SenderId   kol.Id
	Recipients map[dip.Nation]bool

	Body string

	CreatedAt time.Time
	UpdatedAt time.Time
}
