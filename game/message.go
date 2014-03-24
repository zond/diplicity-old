package game

import (
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"regexp"
	"sort"
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
	Id           kol.Id
	GameId       kol.Id `kol:"index"`
	SenderId     kol.Id
	RecipientIds map[string]bool
	SeenBy       map[string]bool

	Body string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (self *Message) Created(d *kol.DB) {
	g := Game{Id: self.GameId}
	if err := d.Get(&g); err != nil {
		panic(err)
	}
	d.EmitUpdate(&g)
}

func (self *Message) EmailTo(c common.SkinnyContext, game *Game, sender *Member, senderUser *user.User, recip *Member, recipUser *user.User, subject string) {
	mailTag := &MailTag{
		M: self.Id,
		R: recip.Id,
	}
	mailTag.H = mailTag.Hash(c.Secret())
	encodedMailTag, err := mailTag.Encode()
	if err != nil {
		c.Errorf("Failed to encode %+v: %v", mailTag, err)
		return
	}

	unsubTag := &common.UnsubscribeTag{
		T: common.UnsubscribeMessageEmail,
		U: recipUser.Id,
	}
	unsubTag.H = unsubTag.Hash(c.Secret())
	encodedUnsubTag, err := unsubTag.Encode()
	if err != nil {
		c.Errorf("Failed to encode %+v: %v", unsubTag, err)
	}

	parts := strings.Split(c.MailAddress(), "@")
	if len(parts) != 2 {
		if c.Env() == common.Development {
			parts = []string{"user", "host.tld"}
		} else {
			c.Errorf("Failed parsing %#v as an email address", c.MailAddress())
			return
		}
	}
	senderName := sender.ShortName(game, senderUser)
	from := fmt.Sprintf("%v <%v+%v@%v>", senderName, parts[0], encodedMailTag, parts[1])
	to := fmt.Sprintf("%v <%v>", recip.Nation, recipUser.Email)
	memberIds := []string{}
	for memberId, _ := range self.RecipientIds {
		memberIds = append(memberIds, memberId)
	}
	sort.Sort(sort.StringSlice(memberIds))
	contextLink, err := recipUser.I("To see this message in context: http://%v/games/%v/messages/%v", recipUser.DiplicityHost, self.GameId, strings.Join(memberIds, "-"))
	if err != nil {
		c.Errorf("Failed translating context link: %v", err)
		return
	}
	unsubLink, err := recipUser.I("To unsubscribe: http://%v/unsubscribe/%v", recipUser.DiplicityHost, encodedUnsubTag)
	if err != nil {
		c.Errorf("Failed translating unsubscribe link: %v", err)
		return
	}
	body := fmt.Sprintf(common.EmailTemplate, self.Body, contextLink, unsubLink)
	if c.Env() == common.Development {
		c.Infof("Would have sent\nFrom: %#v\nTo: %#v\nSubject: %#v\n%v", from, to, subject, body)
	} else {
		if err := c.SendMail(from, subject, body, to); err == nil {
			c.Infof("Sent\nFrom: %#v\nTo: %#v\nSubject: %#v\n%v", from, to, subject, body)
		} else {
			c.Errorf("Unable to send %#v/%#v from %#v to %#v: %v", subject, body, from, to, err)
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
	text := gmail.DecodeText(msg.Text, msg.GetHeader("Content-Type"))
	c.Debugf("Incoming mail to %#v\n%v", msg.GetHeader("To"), text)
	if match := gmail.AddrReg.FindString(msg.GetHeader("To")); match != "" {
		lines := []string{}
		for _, line := range strings.Split(text, "\n") {
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
						Body:         strings.TrimSpace(strings.Join(lines, "\n")),
						GameId:       game.Id,
						RecipientIds: parent.RecipientIds,
					}
					c.Infof("Mail resulted in %+v from %+v", message, sender)
					return message.Send(c, game, sender)
				}
			}
		}
	}
	return nil
}

func (self *Message) Send(c common.SkinnyContext, game *Game, sender *Member) (err error) {
	// make sure the sender is correct
	self.SenderId = sender.Id

	senderUser := &user.User{Id: sender.UserId}
	if err = c.DB().Get(senderUser); err != nil {
		return
	}

	// make sure the sender is one of the recipients
	self.RecipientIds[sender.Id.String()] = true

	// The sender but nobody else saw it...
	self.SeenBy = map[string]bool{
		sender.Id.String(): true,
	}

	// See what phase type the game is in
	var phaseType dip.PhaseType
	switch game.State {
	case common.GameStateCreated:
		phaseType = common.BeforeGamePhaseType
	case common.GameStateStarted:
		var phase *Phase
		if phase, err = game.LastPhase(c.DB()); err != nil {
			return
		}
		phaseType = phase.Type
	case common.GameStateEnded:
		phaseType = common.AfterGamePhaseType
	default:
		err = fmt.Errorf("Unknown game state for %+v", game)
		return
	}

	// Find what chats are allowed during this phase type
	allowedFlags := game.ChatFlags[phaseType]

	// See if the recipient count is allowed
	recipients := len(self.RecipientIds)
	if recipients == 2 {
		if (allowedFlags & common.ChatPrivate) == 0 {
			err = IllegalMessageError{
				Description: fmt.Sprintf("%+v does not allow %+v during %+v", game, self, phaseType),
				Phrase:      "This kind of message is not allowed at this stage of the game",
			}
			return
		}
	} else if recipients == len(common.VariantMap[game.Variant].Nations) {
		if (allowedFlags & common.ChatConference) == 0 {
			err = IllegalMessageError{
				Description: fmt.Sprintf("%+v does not allow %+v during %+v", game, self, phaseType),
				Phrase:      "This kind of message is not allowed at this stage of the game",
			}
			return
		}
	} else if recipients > 2 {
		if (allowedFlags & common.ChatGroup) == 0 {
			err = IllegalMessageError{
				Description: fmt.Sprintf("%+v does not allow %+v during %+v", game, self, phaseType),
				Phrase:      "This kind of message is not allowed at this stage of the game",
			}
			return
		}
	} else {
		err = fmt.Errorf("%+v doesn't have any recipients", self)
		return
	}

	members, err := game.Members(c.DB())
	if err != nil {
		return
	}
	if err = c.DB().Set(self); err != nil {
		return
	}

	for memberId, _ := range self.RecipientIds {
		for _, member := range members {
			if memberId == member.Id.String() && self.SenderId.String() != memberId {
				user := &user.User{Id: member.UserId}
				if err = c.DB().Get(user); err != nil {
					return
				}
				if !user.MessageEmailDisabled && !c.IsSubscribing(user.Email, fmt.Sprintf("/games/%v/messages", game.Id)) {
					memberCopy := member
					gameDescription := ""
					if gameDescription, err = game.Describe(c, user); err != nil {
						return
					}
					go self.EmailTo(c, game, sender, senderUser, &memberCopy, user, gameDescription)
				}
			}
		}
	}

	return
}
