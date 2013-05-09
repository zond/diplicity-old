package web

import (
	"appengine"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/template"
	"time"
	"translation"
)

var htmlTemplates = template.Must(template.New("htmlTemplates").ParseGlob("templates/html/*.html"))
var textTemplates = template.Must(template.New("textTemplates").ParseGlob("templates/text/*"))
var jsTemplates = template.Must(template.New("jsTemplates").ParseGlob("templates/js/*.js"))
var cssTemplates = template.Must(template.New("cssTemplates").ParseGlob("templates/css/*.css"))

type requestData struct {
	response     http.ResponseWriter
	request      *http.Request
	context      appengine.Context
	translations map[string]string
}

func getRequestData(w http.ResponseWriter, r *http.Request) (result requestData) {
	result = requestData{
		response:     w,
		request:      r,
		context:      appengine.NewContext(r),
		translations: translation.GetTranslations(r),
	}
	return
}

func (self requestData) I(phrase string, args ...string) string {
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

func (self requestData) Version() string {
	if appengine.IsDevAppServer() {
		if debugVersion.Before(time.Now().Add(-time.Second)) {
			debugVersion = time.Now()
		}
		return fmt.Sprintf("%v.%v", appengine.VersionID(self.context), debugVersion.UnixNano())
	}
	return appengine.VersionID(self.context)
}

func (self requestData) Inline(p string) string {
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

func renderText(w http.ResponseWriter, r *http.Request, templates *template.Template, template string, data interface{}) {
	if err := templates.ExecuteTemplate(w, template, data); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
}
