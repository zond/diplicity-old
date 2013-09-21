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
	response     http.ResponseWriter
	request      *http.Request
	session      *sessions.Session
	translations map[string]string
	web          *Web
}

func (self *Web) GetRequestData(w http.ResponseWriter, r *http.Request) (result RequestData) {
	result = RequestData{
		response:     w,
		request:      r,
		web:          self,
		translations: translation.GetTranslations(common.GetLanguage(r)),
	}
	result.session, _ = self.sessionStore.Get(r, SessionName)
	return
}

func (self RequestData) Close() {
	self.session.Save(self.request, self.response)
}

func (self RequestData) Appcache() bool {
	return self.web.appcache
}

func (self RequestData) AllocationMethods() (result common.AllocationMethods) {
	for _, meth := range common.AllocationMethodMap {
		meth.Translation = self.I(meth.Name)
		result = append(result, meth)
	}
	sort.Sort(result)
	return
}

func (self RequestData) DefaultAllocationMethod() string {
	return common.RandomString
}

func (self RequestData) DefaultVariant() string {
	return common.StandardString
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
	return url.QueryEscape(common.MustParseURL("http://" + self.request.Host + path).String())
}

func (self RequestData) I(phrase string, args ...string) string {
	pattern, ok := self.translations[phrase]
	if !ok {
		panic(fmt.Errorf("Found no translation for %v", phrase))
	}
	if len(args) > 0 {
		return fmt.Sprintf(pattern, args)
	}
	return pattern
}

func (self RequestData) LogLevel() int {
	return self.web.logLevel
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
	if err := self.web.svgTemplates.ExecuteTemplate(b, p, self); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
	return string(b.Bytes())
}
