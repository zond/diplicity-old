package game

import (
	"crypto/sha1"
	"time"

	"github.com/zond/diplicity/common"
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

type MailTag struct {
	M kol.Id
	D kol.Id
	H []byte
}

func (self *MailTag) Hash() []byte {
	h := sha1.New()
	h.Write(self.M)
	h.Write(self.D)
	return h.Sum(nil)
}

func (self *Message) EmailTo(c common.Context, recip string) {
}
