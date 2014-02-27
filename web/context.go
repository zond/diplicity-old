package web

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
	"github.com/zond/diplicity/common"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/wsubs/gosubs"
)

type Context struct {
	response     http.ResponseWriter
	request      *http.Request
	session      *sessions.Session
	translations map[string]string
	vars         map[string]string
	web          *Web
	db           *kol.DB
}

func (self *Context) DB() *kol.DB {
	return self.db
}

func (self *Context) SetContentType(t string, cache bool) {
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

func (self *Context) RenderJSON(i interface{}) (err error) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return
	}
	self.SetContentType("application/json; charset=UTF-8", false)
	_, err = self.Resp().Write(b)
	return
}

func (self *Context) RenderJS(template string) {
	if err := self.web.jsTemplates.ExecuteTemplate(self.Resp(), template, self); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
	fmt.Fprintln(self.Resp(), ";")
}

func (self *Context) RenderText(templates *template.Template, template string) {
	if err := templates.ExecuteTemplate(self.Resp(), template, self); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
}

func (self *Context) Vars() map[string]string {
	return self.vars
}

func (self *Context) Resp() http.ResponseWriter {
	return self.response
}

func (self *Context) Req() *http.Request {
	return self.request
}

func (self *Context) Close() {
	self.session.Save(self.request, self.response)
}

func (self *Context) Appcache() bool {
	return self.web.appcache
}

func (self *Context) AllocationMethodMap() string {
	result := map[string]common.AllocationMethod{}
	for _, meth := range common.AllocationMethodMap {
		meth.Translation = self.I(meth.Name)
		result[meth.Id] = meth
	}
	return gosubs.Prettify(result)
}

func (self *Context) SecrecyTypesMap() string {
	return gosubs.Prettify(map[string]string{
		"SecretEmail":    self.I("Secret email"),
		"SecretNickname": self.I("Secret nickname"),
		"SecretNation":   self.I("Secret nation"),
	})
}

func (self *Context) AllocationMethods() string {
	result := sort.StringSlice{}
	for _, meth := range common.AllocationMethods {
		result = append(result, meth.Id)
	}
	sort.Sort(result)
	return gosubs.Prettify(result)
}

func (self *Context) DefaultAllocationMethod() string {
	return common.RandomString
}

func (self *Context) DefaultVariant() string {
	return common.ClassicalString
}

func (self *Context) Variants() string {
	result := sort.StringSlice{}
	for _, variant := range common.Variants {
		result = append(result, variant.Id)
	}
	sort.Sort(result)
	return gosubs.Prettify(result)
}

func (self *Context) VariantMap() string {
	result := map[string]common.Variant{}
	for _, variant := range common.Variants {
		variant.Translation = self.I(variant.Name)
		result[variant.Id] = variant
	}
	return gosubs.Prettify(result)
}

func (self *Context) ChatFlagMap() string {
	return gosubs.Prettify(map[string]int{
		"ChatPrivate":    common.ChatPrivate,
		"ChatGroup":      common.ChatGroup,
		"ChatConference": common.ChatConference,
	})
}

func (self *Context) VariantColorizableProvincesMap() string {
	result := map[string][]dip.Province{}
	for _, variant := range common.Variants {
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

func (self *Context) VariantMainProvincesMap() string {
	result := map[string][]dip.Province{}
	for _, variant := range common.Variants {
		result[variant.Id] = []dip.Province{}
		for _, prov := range variant.Graph.Provinces() {
			if prov.Super() == prov {
				result[variant.Id] = append(result[variant.Id], prov)
			}
		}
	}
	return gosubs.Prettify(result)
}

func (self *Context) ChatFlagOptions() (result []common.ChatFlagOption) {
	for _, option := range common.ChatFlagOptions {
		result = append(result, common.ChatFlagOption{
			Id:          option.Id,
			Translation: self.I(option.Name),
		})
	}
	return
}

func (self *Context) Authenticated() bool {
	return true
}

func (self *Context) Abs(path string) string {
	return url.QueryEscape(fmt.Sprintf("http://%v%v", self.request.Host, path))
}

func (self *Context) I(phrase string, args ...string) string {
	pattern, ok := self.translations[phrase]
	if !ok {
		panic(fmt.Errorf("Found no translation for %v", phrase))
	}
	if len(args) > 0 {
		return fmt.Sprintf(pattern, args)
	}
	return pattern
}

func (self *Context) LogLevel() int {
	return self.web.logLevel
}

func (self *Context) GameState(s string) common.GameState {
	switch s {
	case "Created":
		return common.GameStateCreated
	case "Started":
		return common.GameStateStarted
	case "Ended":
		return common.GameStateEnded
	}
	panic(fmt.Errorf("Unknown game state %v", s))
}

func (self *Context) SecretFlagMap() string {
	return gosubs.Prettify(map[string]int{
		"BeforeGame": common.SecretBeforeGame,
		"DuringGame": common.SecretDuringGame,
		"AfterGame":  common.SecretAfterGame,
	})
}

func (self *Context) SecretFlag(s string) common.SecretFlag {
	switch s {
	case "BeforeGame":
		return common.SecretBeforeGame
	case "DuringGame":
		return common.SecretDuringGame
	case "AfterGame":
		return common.SecretAfterGame
	}
	panic(fmt.Errorf("Unknown secret flag %v", s))
}

func (self *Context) ChatFlag(s string) string {
	var rval common.ChatFlag
	switch s {
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

func (self *Context) Version() string {
	return fmt.Sprintf("%v", version.UnixNano())
}

func (self *Context) SVG(p string) string {
	b := new(bytes.Buffer)
	if err := self.web.svgTemplates.ExecuteTemplate(b, p, self); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
	return string(b.Bytes())
}
