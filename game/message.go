package game

import (
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jhillyerd/go.enmime"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	"github.com/zond/gmail"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
)

var emailPlusReg = regexp.MustCompile("^.+\\+(.+)@.+$")

type MailTag struct {
	M kol.Id
	R kol.Id
	H []byte
}

func (self *MailTag) Hash(secret string) []byte {
	h := sha1.New()
	h.Write(self.M)
	h.Write(self.R)
	h.Write([]byte(secret))
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

func DecodeMailTag(secret string, s string) (result *MailTag, err error) {
	buf := bytes.NewBufferString(s)
	dec := gob.NewDecoder(base64.NewDecoder(base64.URLEncoding, buf))
	tag := &MailTag{}
	if err = dec.Decode(tag); err != nil {
		return
	}
	wanted := tag.Hash(secret)
	if len(wanted) != len(tag.H) || subtle.ConstantTimeCompare(wanted, tag.H) != 1 {
		err = fmt.Errorf("%+v has wrong hash, wanted %v", tag, wanted)
		return
	}
	result = tag
	return
}

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

func (self *Message) EmailTo(c common.SkinnyContext, sender, recip *Member, recipUser *user.User, subject string) {
	tag := &MailTag{
		M: self.Id,
		R: recip.Id,
	}
	tag.H = tag.Hash(c.Secret())
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

type IllegalMessageError struct {
	Description string
	Phrase      string
}

func (self IllegalMessageError) Error() string {
	return self.Description
}

func IncomingMail(c common.SkinnyContext, msg *enmime.MIMEBody) (err error) {
	c.Infof("Incoming mail to %#v\n%v", msg.GetHeader("To"), msg.Text)
	if match := gmail.AddrReg.FindString(msg.GetHeader("To")); match != "" {
		lines := []string{}
		for _, line := range strings.Split(msg.Text, "\n") {
			if !strings.Contains(line, c.MailAddress()) && strings.Index(line, ">") != 0 {
				lines = append(lines, line)
			}
		}
		for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
			lines = lines[1:]
		}
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
		if len(lines) > 0 {
			if match2 := emailPlusReg.FindStringSubmatch(match); match2 != nil {
				var tag *MailTag
				if tag, err = DecodeMailTag(c.Secret(), match2[1]); err == nil {
					sender := &Member{Id: tag.R}
					if err = c.DB().Get(sender); err != nil {
						return
					}
					parent := &Message{Id: tag.M}
					if err = c.DB().Get(parent); err != nil {
						return
					}
					game := &Game{Id: parent.GameId}
					if err = c.DB().Get(game); err != nil {
						return
					}
					message := &Message{
						Body:       strings.Join(lines, "\n"),
						GameId:     game.Id,
						Recipients: parent.Recipients,
					}
					return SendMessage(c, game, sender, message)
				}
			}
		}
	}
	return nil
}

func SendMessage(c common.SkinnyContext, game *Game, sender *Member, message *Message) (err error) {
	// make sure the sender is correct
	message.SenderId = sender.Id

	// make sure the sender is one of the recipients
	message.Recipients[sender.Nation] = true

	var phaseDescription func(u *user.User) string
	phaseDescriptionParams := []string{}
	var phaseType dip.PhaseType
	switch game.State {
	case common.GameStateCreated:
		phaseType = common.BeforeGamePhaseType
		phaseDescription = func(u *user.User) string { return string(phaseType) }
		phaseDescriptionParams = []string{}
	case common.GameStateStarted:
		var phase *Phase
		if phase, err = game.LastPhase(c.DB()); err != nil {
			return
		}
		phaseType = phase.Type
		phaseDescription = func(u *user.User) (result string) {
			result, _ = u.I("game_phase_description", fmt.Sprint(phase.Year))
			return
		}
		phaseDescriptionParams = []string{string(phase.Season), string(phase.Type)}
	case common.GameStateEnded:
		phaseType = common.AfterGamePhaseType
		phaseDescription = func(u *user.User) string { return string(phaseType) }
		phaseDescriptionParams = []string{}
	default:
		err = fmt.Errorf("Unknown game state for %+v", game)
		return
	}

	allowedFlags := game.ChatFlags[phaseType]

	recipients := len(message.Recipients)
	if recipients == 2 {
		if (allowedFlags & common.ChatPrivate) == 0 {
			err = IllegalMessageError{
				Description: fmt.Sprintf("%+v does not allow %+v during %+v", game, message, phaseType),
				Phrase:      "This kind of message is not allowed at this stage of the game",
			}
			return
		}
	} else if recipients == len(common.VariantMap[game.Variant].Nations) {
		if (allowedFlags & common.ChatConference) == 0 {
			err = IllegalMessageError{
				Description: fmt.Sprintf("%+v does not allow %+v during %+v", game, message, phaseType),
				Phrase:      "This kind of message is not allowed at this stage of the game",
			}
			return
		}
	} else if recipients > 2 {
		if (allowedFlags & common.ChatGroup) == 0 {
			err = IllegalMessageError{
				Description: fmt.Sprintf("%+v does not allow %+v during %+v", game, message, phaseType),
				Phrase:      "This kind of message is not allowed at this stage of the game",
			}
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
	if err = c.DB().Set(message); err != nil {
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
						go message.EmailTo(c, sender, &memberCopy, user, fmt.Sprintf(phaseDescription(user), parts...))
					}
				}
			}
		}
	}

	return
}
