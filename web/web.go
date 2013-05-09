package web

import (
	"fmt"
	"net/http"
	"text/template"
)

var htmlTemplates = template.Must(template.New("htmlTemplates").ParseGlob("templates/html/*.html"))
var textTemplates = template.Must(template.New("textTemplates").ParseGlob("templates/text/*"))
var jsModelTemplates = template.Must(template.New("jsCollectionTemplates").ParseGlob("templates/js/models/*.js"))
var jsCollectionTemplates = template.Must(template.New("jsModelTemplates").ParseGlob("templates/js/collections/*.js"))
var jsTemplates = template.Must(template.New("jsTemplates").ParseGlob("templates/js/*.js"))
var cssTemplates = template.Must(template.New("cssTemplates").ParseGlob("templates/css/*.css"))

func renderText(w http.ResponseWriter, r *http.Request, templates *template.Template, template string, data interface{}) {
	if err := templates.ExecuteTemplate(w, template, data); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
}
