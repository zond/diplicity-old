package common

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"time"
)

type RequestData struct {
	Response     http.ResponseWriter
	Request      *http.Request
	translations map[string]string
}

func GetRequestData(w http.ResponseWriter, r *http.Request) (result RequestData) {
	result = RequestData{
		Response:     w,
		Request:      r,
		translations: getTranslations(r),
	}
	return
}

func (self RequestData) Variants() (result Variants) {
	for _, variant := range VariantMap {
		variant.Translation = self.I(variant.Name)
		result = append(result, variant)
	}
	sort.Sort(result)
	return
}

func (self RequestData) ChatFlagOptions() (result []ChatFlagOption) {
	for _, option := range ChatFlagOptions {
		result = append(result, ChatFlagOption{
			Id:          option.Id,
			Translation: self.I(option.Name),
		})
	}
	return
}

func (self RequestData) Authenticated() bool {
	return true
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

func (self RequestData) ChatFlag(s string) string {
	var rval ChatFlag
	switch s {
	case "White":
		rval = ChatWhite
	case "Grey":
		rval = ChatGrey
	case "Black":
		rval = ChatBlack
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
