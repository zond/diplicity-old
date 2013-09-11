package web

import (
	"bytes"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/zond/kcwraps/kol"
	"net/http"
	"regexp"
	"strings"
	"text/template"
)

const (
	SessionEmail = "email"
	SessionName  = "diplicity_session"
)

type Web struct {
	sessionStore *sessions.CookieStore
	db           *kol.DB
	env          string
}

func New(env, secret string) (result *Web) {
	result = &Web{
		env:          env,
		sessionStore: sessions.NewCookieStore([]byte(secret)),
	}
	var err error
	if result.db, err = kol.New("diplicity"); err != nil {
		panic(err)
	}
	return
}

var spaceRegexp = regexp.MustCompile("\\s+")

var svgTemplates = template.Must(template.New("svgTemplates").ParseGlob("templates/svg/*.svg"))
var htmlTemplates = template.Must(template.New("htmlTemplates").ParseGlob("templates/html/*.html"))
var textTemplates = template.Must(template.New("textTemplates").ParseGlob("templates/text/*"))
var jsModelTemplates = template.Must(template.New("jsCollectionTemplates").ParseGlob("templates/js/models/*.js"))
var jsCollectionTemplates = template.Must(template.New("jsModelTemplates").ParseGlob("templates/js/collections/*.js"))
var jsTemplates = template.Must(template.New("jsTemplates").ParseGlob("templates/js/*.js"))
var cssTemplates = template.Must(template.New("cssTemplates").ParseGlob("templates/css/*.css"))
var _Templates = template.Must(template.New("_Templates").ParseGlob("templates/_/*.html"))
var jsViewTemplates = template.Must(template.New("jsViewTemplates").ParseGlob("templates/js/views/*.js"))

func (self *Web) renderText(w http.ResponseWriter, r *http.Request, templates *template.Template, template string, data interface{}) {
	if err := templates.ExecuteTemplate(w, template, data); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
}

func (self *Web) render_Templates(data RequestData) {
	fmt.Fprintln(data.Response, "(function() {")
	fmt.Fprintln(data.Response, "  var n;")
	var buf *bytes.Buffer
	var rendered string
	for _, templ := range _Templates.Templates() {
		fmt.Fprintf(data.Response, "  n = $('<script type=\"text/template\" id=\"%v_underscore\"></script>');\n", strings.Split(templ.Name(), ".")[0])
		fmt.Fprintf(data.Response, "  n.text('")
		buf = new(bytes.Buffer)
		templ.Execute(buf, data)
		rendered = string(buf.Bytes())
		rendered = spaceRegexp.ReplaceAllString(rendered, " ")
		rendered = strings.Replace(rendered, "\\", "\\\\", -1)
		rendered = strings.Replace(rendered, "'", "\\'", -1)
		fmt.Fprint(data.Response, rendered)
		fmt.Fprintln(data.Response, "');")
		fmt.Fprintln(data.Response, "  $('head').append(n);")
	}
	fmt.Fprintln(data.Response, "})();")
}
