package game

import (
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"strings"
	"time"

	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
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
	R kol.Id
	H []byte
}

func (self *MailTag) Hash() []byte {
	h := sha1.New()
	h.Write(self.M)
	h.Write(self.R)
	return h.Sum(nil)
}

func (self *MailTag) Encode() (result string, err error) {
	buf := &bytes.Buffer{}
	baseEnc := base64.NewEncoder(base64.URLEncoding, buf)
	gobEnc := gob.NewEncoder(baseEnc)
	if err = gobEnc.Encode(self); err != nil {
		return
	}
	if err = baseEnc.Close(); err != nil {
		return
	}
	result = buf.String()
	return
}

func DecodeMailTag(s string) (result *MailTag, err error) {
	buf := bytes.NewBufferString(s)
	dec := gob.NewDecoder(base64.NewDecoder(base64.URLEncoding, buf))
	tag := &MailTag{}
	if err = dec.Decode(tag); err != nil {
		return
	}
	wanted := tag.Hash()
	if len(wanted) != len(tag.H) || subtle.ConstantTimeCompare(wanted, tag.H) != 1 {
		err = fmt.Errorf("%+v has wrong hash, wanted %v", tag, wanted)
		return
	}
	result = tag
	return
}

func (self *Message) EmailTo(c common.WSContext, sender, recip *Member, recipUser *user.User, subject string) {
	tag := &MailTag{
		M: self.Id,
		R: recip.Id,
	}
	tag.H = tag.Hash()
	encoded, err := tag.Encode()
	if err != nil {
		c.Errorf("Failed to encode %+v: %v", tag, err)
		return
	}
	parts := strings.Split(c.MailAddress(), "@")
	if len(parts) != 2 {
		c.Errorf("Failed parsing %#v as an email address", c.MailAddress())
		return
	}
	from := fmt.Sprintf("%v <%v+%v@%v>", sender.Nation, parts[0], encoded, parts[1])
	to := fmt.Sprintf("%v <%v>", recip.Nation, recipUser.Email)
	if c.Env() == "development" {
		c.Infof("Would have sent\nFrom: %#v\nTo: %#v\nSubject: %#v\n%v", from, to, subject, self.Body)
	} else {
		if err := c.SendMail(from, subject, self.Body, to); err == nil {
			c.Infof("Sent\nFrom: %#v\nTo: %#v\nSubject: %#v\n%v", from, to, subject, self.Body)
		} else {
			c.Errorf("Unable to send %#v/%#v from %#v to %#v: %v", subject, self.Body, from, to, err)
		}
	}
}
