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
	Id        kol.Id
	ChannelId kol.Id
	GameId    kol.Id
	SenderId  kol.Id

	Senders map[dip.Nation]dip.Nation

	Body string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Channel struct {
	Id     kol.Id
	GameId kol.Id

	RealMembers    map[dip.Nation]bool
	VirtualMembers map[dip.Nation]map[dip.Nation]bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
