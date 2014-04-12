package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/wsubs/gosubs"
)

type HTTPContext struct {
	response     http.ResponseWriter
	request      *http.Request
	session      *sessions.Session
	translations map[string]string
	vars         map[string]string
	web          *Web
}

func (self *HTTPContext) Env() string {
	return self.web.Env()
}

func (self *HTTPContext) DB() *kol.DB {
	return self.web.db
}

func (self *HTTPContext) Secret() string {
	return self.web.secret
}

func (self *HTTPContext) Session() *sessions.Session {
	return self.session
}

func (self *HTTPContext) SetContentType(t string, cache bool) {
	self.Resp().Header().Set("Content-Type", t)
	self.Resp().Header().Set("Vary", "Accept")
	if cache {
		self.Resp().Header().Set("Cache-Control", "public, max-age=864000")
	} else {
		self.Resp().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		self.Resp().Header().Set("Pragma", "no-cache")
		self.Resp().Header().Set("Expires", "0")
	}
}

func (self *HTTPContext) Fatalf(format string, args ...interface{}) {
	self.web.Fatalf(format, args...)
}

func (self *HTTPContext) Errorf(format string, args ...interface{}) {
	self.web.Errorf(format, args...)
}

func (self *HTTPContext) Infof(format string, args ...interface{}) {
	self.web.Infof(format, args...)
}

func (self *HTTPContext) Debugf(format string, args ...interface{}) {
	self.web.Debugf(format, args...)
}

func (self *HTTPContext) Tracef(format string, args ...interface{}) {
	self.web.Tracef(format, args...)
}

func (self *HTTPContext) RenderJSON(i interface{}) (err error) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return
	}
	self.SetContentType("application/json; charset=UTF-8", false)
	_, err = self.Resp().Write(b)
	return
}

func (self *HTTPContext) RenderJS(template string) {
	if err := self.web.jsTemplates.ExecuteTemplate(self.Resp(), template, self); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
	fmt.Fprintln(self.Resp(), ";")
}

func (self *HTTPContext) RenderText(templates *template.Template, template string) {
	if err := templates.ExecuteTemplate(self.Resp(), template, self); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
}

func (self *HTTPContext) Vars() map[string]string {
	return self.vars
}

func (self *HTTPContext) Resp() http.ResponseWriter {
	return self.response
}

func (self *HTTPContext) Req() *http.Request {
	return self.request
}

func (self *HTTPContext) Close() {
	self.session.Save(self.request, self.response)
}

func (self *HTTPContext) Appcache() bool {
	if self.Req().URL.Path != "" && self.Req().URL.Path != "/" {
		return false
	}
	return self.web.appcache
}

func (self *HTTPContext) mI(pattern string, args ...interface{}) (result string) {
	result, _ = self.I(pattern, args...)
	return
}

func (self *HTTPContext) AllocationMethodMap() string {
	result := map[string]AllocationMethod{}
	for _, meth := range AllocationMethodMap {
		meth.Translation = self.mI(meth.Name)
		result[meth.Id] = meth
	}
	return gosubs.Prettify(result)
}

func (self *HTTPContext) SecrecyTypesMap() string {
	return gosubs.Prettify(map[string]string{
		"SecretEmail":    self.mI("Secret email"),
		"SecretNickname": self.mI("Secret nickname"),
		"SecretNation":   self.mI("Secret nation"),
	})
}

func (self *HTTPContext) AllocationMethods() string {
	result := sort.StringSlice{}
	for _, meth := range AllocationMethods {
		result = append(result, meth.Id)
	}
	sort.Sort(result)
	return gosubs.Prettify(result)
}

func (self *HTTPContext) DefaultAllocationMethod() string {
	return RandomString
}

func (self *HTTPContext) DefaultVariant() string {
	return ClassicalString
}

func (self *HTTPContext) Variants() string {
	result := sort.StringSlice{}
	for _, variant := range Variants {
		result = append(result, variant.Id)
	}
	sort.Sort(result)
	return gosubs.Prettify(result)
}

func (self *HTTPContext) VariantMap() string {
	result := map[string]Variant{}
	for _, variant := range Variants {
		variant.Translation = self.mI(variant.Name)
		result[variant.Id] = variant
	}
	return gosubs.Prettify(result)
}

func (self *HTTPContext) ConsequenceMap() string {
	return gosubs.Prettify(map[string]Consequence{
		"ReliabilityHit": ReliabilityHit,
		"NoWait":         NoWait,
		"Surrender":      Surrender,
	})
}

func (self *HTTPContext) ChatFlagMap() string {
	return gosubs.Prettify(map[string]int{
		"ChatPrivate":    ChatPrivate,
		"ChatGroup":      ChatGroup,
		"ChatConference": ChatConference,
	})
}

func (self *HTTPContext) VariantColorizableProvincesMap() string {
	result := map[string][]dip.Province{}
	for _, variant := range Variants {
		supers := map[dip.Province]bool{}
		for _, prov := range variant.Graph.Provinces() {
			supers[prov.Super()] = true
		}
		for prov, _ := range supers {
			result[variant.Id] = append(result[variant.Id], prov)
		}
	}
	return gosubs.Prettify(result)
}

func (self *HTTPContext) VariantMainProvincesMap() string {
	result := map[string][]dip.Province{}
	for _, variant := range Variants {
		result[variant.Id] = []dip.Province{}
		for _, prov := range variant.Graph.Provinces() {
			if prov.Super() == prov {
				result[variant.Id] = append(result[variant.Id], prov)
			}
		}
	}
	return gosubs.Prettify(result)
}

func (self *HTTPContext) ConsequenceOptions() (result []ConsequenceOption) {
	for _, option := range ConsequenceOptions {
		result = append(result, ConsequenceOption{
			Id:          option.Id,
			Translation: self.mI(option.Name),
		})
	}
	return
}

func (self *HTTPContext) ChatFlagOptions() (result []ChatFlagOption) {
	for _, option := range ChatFlagOptions {
		result = append(result, ChatFlagOption{
			Id:          option.Id,
			Translation: self.mI(option.Name),
		})
	}
	return
}

func (self *HTTPContext) Authenticated() bool {
	return true
}

func (self *HTTPContext) Abs(path string) string {
	return url.QueryEscape(fmt.Sprintf("http://%v%v", self.request.Host, path))
}

func (self *HTTPContext) I(phrase string, args ...interface{}) (result string, err error) {
	pattern, ok := self.translations[phrase]
	if !ok {
		err = fmt.Errorf("Found no translation for %v", phrase)
		result = err.Error()
		return
	}
	if len(args) > 0 {
		result = fmt.Sprintf(pattern, args...)
		return
	}
	result = pattern
	return
}

func (self *HTTPContext) LogLevel() int {
	return self.web.logLevel
}

func (self *HTTPContext) GameState(s string) (result GameState, err error) {
	switch s {
	case "Created":
		result = GameStateCreated
		return
	case "Started":
		result = GameStateStarted
		return
	case "Ended":
		result = GameStateEnded
		return
	}
	err = fmt.Errorf("Unknown game state %v", s)
	return
}

func (self *HTTPContext) SecretFlagMap() string {
	return gosubs.Prettify(map[string]int{
		"BeforeGame": SecretBeforeGame,
		"DuringGame": SecretDuringGame,
		"AfterGame":  SecretAfterGame,
	})
}

func (self *HTTPContext) Consequence(s string) Consequence {
	switch s {
	case "ReliabilityHit":
		return ReliabilityHit
	case "NoWait":
		return NoWait
	case "Surrender":
		return Surrender
	}
	panic(fmt.Errorf("Unknown consequence flag %v", s))
}

func (self *HTTPContext) SecretFlag(s string) SecretFlag {
	switch s {
	case "BeforeGame":
		return SecretBeforeGame
	case "DuringGame":
		return SecretDuringGame
	case "AfterGame":
		return SecretAfterGame
	}
	panic(fmt.Errorf("Unknown secret flag %v", s))
}

func (self *HTTPContext) ChatFlag(s string) string {
	var rval ChatFlag
	switch s {
	case "Private":
		rval = ChatPrivate
	case "Group":
		rval = ChatGroup
	case "Conference":
		rval = ChatConference
	}
	return fmt.Sprint(rval)
}

var version = time.Now()

func (self *HTTPContext) Version() string {
	return fmt.Sprintf("%v", version.UnixNano())
}

func (self *HTTPContext) SVG(p string) string {
	b := new(bytes.Buffer)
	if err := self.web.svgTemplates.ExecuteTemplate(b, p, self); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
	return string(b.Bytes())
}
