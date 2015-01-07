package unsubscribe

import (
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/zond/unbolted"
)

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
	U unbolted.Id
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
