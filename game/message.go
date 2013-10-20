package game

import (
	"github.com/zond/diplicity/common"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
	"time"
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
	Id     kol.Id
	GameId kol.Id `kol:"index"`

	Sender  kol.Id
	Nation  dip.Nation
	Channel common.ChatChannel

	Body string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (self *Message) sender(d *kol.DB) (result *Member, err error) {
	result = &Member{
		Id: self.Sender,
	}
	err = d.Get(result)
	return
}
