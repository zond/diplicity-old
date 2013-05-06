package web

import (
	"appengine"
	"fmt"
	htmlTemplate "html/template"
	"net/http"
	textTemplate "text/template"
)

var htmlTemplates = htmlTemplate.Must(htmlTemplate.New("htmlTemplates").ParseGlob("templates/html/*.html"))
var jsTemplates = textTemplate.Must(textTemplate.New("jsTemplates").ParseGlob("templates/js/*.js"))
var cssTemplates = textTemplate.Must(textTemplate.New("cssTemplates").ParseGlob("templates/css/*.css"))

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
	return appengine.VersionID(self.context)
}

func renderHtml(w http.ResponseWriter, r *http.Request, templates *htmlTemplate.Template, template string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	if err := templates.ExecuteTemplate(w, template, data); err != nil {
		panic(fmt.Errorf("While rendering HTML: %v", err))
	}
}

func renderText(w http.ResponseWriter, r *http.Request, templates *textTemplate.Template, template string, data interface{}) {
	if err := templates.ExecuteTemplate(w, template, data); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
}
