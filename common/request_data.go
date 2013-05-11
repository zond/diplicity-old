package common

import (
	"appengine"
	"appengine/user"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"time"
)

type RequestData struct {
	Response     http.ResponseWriter
	Request      *http.Request
	Context      appengine.Context
	User         *user.User
	translations map[string]string
}

func GetRequestData(w http.ResponseWriter, r *http.Request) (result RequestData) {
	result = RequestData{
		Response:     w,
		Request:      r,
		Context:      appengine.NewContext(r),
		translations: getTranslations(r),
	}
	result.User = user.Current(result.Context)
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

func (self RequestData) Authenticated() bool {
	if self.User == nil {
		loginUrl, err := user.LoginURL(self.Context, HostURL(self.Request))
		if err != nil {
			panic(err)
		}
		self.Response.Header().Set("Location", loginUrl)
		self.Response.WriteHeader(401)
		fmt.Fprintln(self.Response, "Unauthorized")
		return false
	}
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

var debugVersion time.Time

func (self RequestData) Version() string {
	if appengine.IsDevAppServer() {
		if debugVersion.Before(time.Now().Add(-time.Second)) {
			debugVersion = time.Now()
		}
		return fmt.Sprintf("%v.%v", appengine.VersionID(self.Context), debugVersion.UnixNano())
	}
	return appengine.VersionID(self.Context)
}

func (self RequestData) Inline(p string) string {
	in, err := os.Open(p)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(in)
	if err != nil {
		panic(err)
	}
	return string(b)
}
