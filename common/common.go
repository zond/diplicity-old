package common

import (
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/wsubs/gosubs"
)

type Translator interface {
	I(phrase string, args ...interface{}) (result string, err error)
}

const (
	DefaultSecret       = "something very very secret"
	SubscriptionTimeout = time.Minute * 15
)

type Mailer interface {
	SendMail(fromName, replyTo, subject, message string, recips []string) error
	ReceiveAddress() string
	SendAddress() string
}

type SkinnyContext interface {
	gosubs.Logger
	gosubs.SubscriptionManager
	Mailer
	DB() *kol.DB
	BetweenTransactions(func(c SkinnyContext))
	Transact(func(c SkinnyContext) error) error
	Env() string
	Secret() string
}

type skinnyWeb struct {
	*Web
	db *kol.DB
}

func (self skinnyWeb) BetweenTransactions(f func(c SkinnyContext)) {
	self.db.BetweenTransactions(func(d *kol.DB) {
		self.db = d
		f(self)
	})
}

func (self skinnyWeb) Transact(f func(c SkinnyContext) error) error {
	return self.db.Transact(func(d *kol.DB) error {
		self.db = d
		return f(self)
	})
}

func (self skinnyWeb) DB() *kol.DB {
	return self.db
}

type skinnyWSContext struct {
	WSContext
}

func (self skinnyWSContext) BetweenTransactions(f func(c SkinnyContext)) {
	self.WSContext.BetweenTransactions(func(c WSContext) {
		f(skinnyWSContext{c})
	})
}

func (self skinnyWSContext) Transact(f func(c SkinnyContext) error) error {
	return self.WSContext.Transact(func(c WSContext) error {
		return f(skinnyWSContext{c})
	})
}

var prefPattern = regexp.MustCompile("^([^\\s;]+)(;q=([\\d.]+))?$")

func MostAccepted(r *http.Request, def, name string) string {
	bestValue := def
	var bestScore float64 = -1
	var score float64
	for _, pref := range strings.Split(r.Header.Get(name), ",") {
		if match := prefPattern.FindStringSubmatch(pref); match != nil {
			score = 1
			if match[3] != "" {
				score, _ = strconv.ParseFloat(match[3], 64)
			}
			if score > bestScore {
				bestScore = score
				bestValue = match[1]
			}
		}
	}
	return bestValue
}

func HostURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%v://%v/reload", scheme, r.Host)
}

const (
	UnsubscribeMessageEmail = iota
	UnsubscribePhaseEmail
)

const (
	EmailTemplate = `%v
----
%v
%v`
)

type UnsubscribeTag struct {
	T int
	U kol.Id
	H []byte
}

func (self *UnsubscribeTag) Hash(secret string) []byte {
	h := sha1.New()
	h.Write(self.U)
	h.Write([]byte(secret))
	return h.Sum(nil)
}

func (self *UnsubscribeTag) Encode() (result string, err error) {
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

func DecodeUnsubscribeTag(secret string, s string) (result *UnsubscribeTag, err error) {
	buf := bytes.NewBufferString(s)
	dec := gob.NewDecoder(base64.NewDecoder(base64.URLEncoding, buf))
	tag := &UnsubscribeTag{}
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
