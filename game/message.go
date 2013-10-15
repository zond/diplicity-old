package game

import (
	"github.com/zond/diplicity/common"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
	"time"
)

type Message struct {
	Id     kol.Id
	GameId kol.Id `kol:"index"`

	Sender        kol.Id
	VirtualSender dip.Nation
	Flag          common.ChatFlag
	Channel       common.ChatChannel

	Body string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (self *Message) toState(d *kol.DB) (result *MessageState, err error) {
	result = &MessageState{
		Message: &Message{
			Body: self.Body,
		},
		Member: &Member{},
	}
	if self.Flag == common.ChatGrey {
		result.Member.Nation = common.Anonymous
	}
	if self.Flag == common.ChatBlack {
		result.Member.Nation = self.VirtualSender
	}
	if self.Flag == common.ChatWhite {
		member := &Member{Id: self.Sender}
		if err = d.Get(member); err != nil {
			return
		}
		result.Member.Nation = member.Nation
	}
	return
}
