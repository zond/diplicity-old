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

func (self *Message) EmailTo(c common.SkinnyContext, sender, recip *Member, recipUser *user.User, subject string) {
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

func SendMessage(c common.SkinnyContext, game *Game, sender *Member, message *Message) (err error) {
	// make sure the sender is correct
	message.SenderId = sender.Id

	// make sure the sender is one of the recipients
	message.Recipients[sender.Nation] = true

	phaseDescription := ""
	phaseDescriptionParams := []string{}
	var phaseType dip.PhaseType
	switch game.State {
	case common.GameStateCreated:
		phaseType = common.BeforeGamePhaseType
		phaseDescription = "%v"
		phaseDescriptionParams = []string{string(phaseType)}
	case common.GameStateStarted:
		var phase *Phase
		if phase, err = game.LastPhase(c.DB()); err != nil {
			return
		}
		phaseType = phase.Type
		phaseDescription = fmt.Sprintf("%%v %v, %%v", phase.Year)
		phaseDescriptionParams = []string{string(phase.Season), string(phase.Type)}
	case common.GameStateEnded:
		phaseType = common.AfterGamePhaseType
		phaseDescription = "%v"
		phaseDescriptionParams = []string{string(phaseType)}
	default:
		err = fmt.Errorf("Unknown game state for %+v", game)
		return
	}

	allowedFlags := game.ChatFlags[phaseType]

	recipients := len(message.Recipients)
	if recipients == 2 {
		if (allowedFlags & common.ChatPrivate) == 0 {
			err = fmt.Errorf("%+v does not allow %+v during %+v (%v)", game, message, phaseType, common.ChatPrivate)
			return
		}
	} else if recipients == len(common.VariantMap[game.Variant].Nations) {
		if (allowedFlags & common.ChatConference) == 0 {
			err = fmt.Errorf("%+v does not allow %+v during %+v", game, message, phaseType)
			return
		}
	} else if recipients > 2 {
		if (allowedFlags & common.ChatGroup) == 0 {
			err = fmt.Errorf("%+v does not allow %+v during %+v", game, message, phaseType)
			return
		}
	} else {
		err = fmt.Errorf("%+v doesn't have any recipients", message)
		return
	}

	members, err := game.Members(c.DB())
	if err != nil {
		return
	}
	if err = c.DB().Set(&message); err != nil {
		return
	}
	if c.MailAddress() != "" {
		for recip, _ := range message.Recipients {
			for _, member := range members {
				if recip == member.Nation && !message.SenderId.Equals(member.Id) {
					user := &user.User{Id: member.UserId}
					if err = c.DB().Get(user); err != nil {
						return
					}
					if !user.MessageEmailDisabled && !c.IsSubscribing(user.Email, fmt.Sprintf("/games/%v/messages", game.Id)) {
						memberCopy := member
						parts := make([]interface{}, len(phaseDescriptionParams))
						for i, param := range phaseDescriptionParams {
							if parts[i], err = user.I(param); err != nil {
								return
							}
						}
						go message.EmailTo(c, sender, &memberCopy, user, fmt.Sprintf(phaseDescription, parts...))
					}
				}
			}
		}
	}

	return
}
