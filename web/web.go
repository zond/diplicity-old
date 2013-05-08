package web

import (
	"appengine"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os"
	"text/template"
	"time"
)

var htmlTemplates = template.Must(template.New("htmlTemplates").ParseGlob("templates/html/*.html"))
var textTemplates = template.Must(template.New("textTemplates").ParseGlob("templates/text/*"))
var jsTemplates = template.Must(template.New("jsTemplates").ParseGlob("templates/js/*.js"))
var cssTemplates = template.Must(template.New("cssTemplates").ParseGlob("templates/css/*.css"))

type requestData struct {
	response http.ResponseWriter
	request  *http.Request
	context  appengine.Context
}

func getRequestData(w http.ResponseWriter, r *http.Request) requestData {
	return requestData{
		response: w,
		request:  r,
		context:  appengine.NewContext(r),
	}
}

func (self requestData) Version() string {
	if os.Getenv("NOCACHE") == "t" && appengine.IsDevAppServer() {
		return fmt.Sprintf("%v.%v", appengine.VersionID(self.context), time.Now().UnixNano())
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
