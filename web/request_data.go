package web

import (
	"bytes"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/translation"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type RequestData struct {
	Response     http.ResponseWriter
	Request      *http.Request
	Session      *sessions.Session
	Translations map[string]string
	Env          string
}

func (self *Web) GetRequestData(w http.ResponseWriter, r *http.Request) (result RequestData) {
	result = RequestData{
		Response:     w,
		Request:      r,
		Env:          self.env,
		Translations: translation.GetTranslations(common.GetLanguage(r)),
	}
	result.Session, _ = self.sessionStore.Get(r, SessionName)
	return
}

func (self RequestData) Close() {
	self.Session.Save(self.Request, self.Response)
}

func (self RequestData) Variants() (result common.Variants) {
	for _, variant := range common.VariantMap {
		variant.Translation = self.I(variant.Name)
		result = append(result, variant)
	}
	sort.Sort(result)
	return
}

func (self RequestData) ChatFlagOptions() (result []common.ChatFlagOption) {
	for _, option := range common.ChatFlagOptions {
		result = append(result, common.ChatFlagOption{
			Id:          option.Id,
			Translation: self.I(option.Name),
		})
	}
	return
}

func (self RequestData) Authenticated() bool {
	return true
}

func (self RequestData) Abs(path string) string {
	return url.QueryEscape(common.MustParseURL("http://" + self.Request.Host + path).String())
}

func (self RequestData) I(phrase string, args ...string) string {
	pattern, ok := self.Translations[phrase]
	if !ok {
		panic(fmt.Errorf("Found no translation for %v", phrase))
	}
	if len(args) > 0 {
		return fmt.Sprintf(pattern, args)
	}
	return pattern
}

func (self RequestData) LogLevel() int {
	return common.LogLevel
}

func (self RequestData) ChatFlag(s string) string {
	var rval common.ChatFlag
	switch s {
	case "White":
		rval = common.ChatWhite
	case "Grey":
		rval = common.ChatGrey
	case "Black":
		rval = common.ChatBlack
	case "Private":
		rval = common.ChatPrivate
	case "Group":
		rval = common.ChatGroup
	case "Conference":
		rval = common.ChatConference
	}
	return fmt.Sprint(rval)
}

var version = time.Now()

func (self RequestData) Version() string {
	return fmt.Sprintf("%v", version.UnixNano())
}

func (self RequestData) SVG(p string) string {
	b := new(bytes.Buffer)
	if err := svgTemplates.ExecuteTemplate(b, p, self); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
	return string(b.Bytes())
}
