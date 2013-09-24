package web

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/zond/kcwraps/kol"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

const (
	SessionEmail  = "email"
	SessionName   = "diplicity_session"
	DefaultSecret = "something very secret"
)

const (
	Fatal = iota
	Error
	Info
	Debug
	Trace
)

var spaceRegexp = regexp.MustCompile("\\s+")

type Web struct {
	sessionStore          *sessions.CookieStore
	db                    *kol.DB
	env                   string
	logLevel              int
	appcache              bool
	svgTemplates          *template.Template
	htmlTemplates         *template.Template
	textTemplates         *template.Template
	jsModelTemplates      *template.Template
	jsCollectionTemplates *template.Template
	jsTemplates           *template.Template
	cssTemplates          *template.Template
	_Templates            *template.Template
	jsViewTemplates       *template.Template
}

func New() (result *Web) {
	result = &Web{
		appcache:              true,
		svgTemplates:          template.Must(template.New("svgTemplates").ParseGlob("templates/svg/*.svg")),
		htmlTemplates:         template.Must(template.New("htmlTemplates").ParseGlob("templates/html/*.html")),
		textTemplates:         template.Must(template.New("textTemplates").ParseGlob("templates/text/*")),
		jsModelTemplates:      template.Must(template.New("jsCollectionTemplates").ParseGlob("templates/js/models/*.js")),
		jsCollectionTemplates: template.Must(template.New("jsModelTemplates").ParseGlob("templates/js/collections/*.js")),
		jsTemplates:           template.Must(template.New("jsTemplates").ParseGlob("templates/js/*.js")),
		cssTemplates:          template.Must(template.New("cssTemplates").ParseGlob("templates/css/*.css")),
		_Templates:            template.Must(template.New("_Templates").ParseGlob("templates/_/*.html")),
		jsViewTemplates:       template.Must(template.New("jsViewTemplates").ParseGlob("templates/js/views/*.js")),
		db:                    kol.Must("diplicity"),
	}
	return
}

func (self *Web) SetEnv(env string) *Web {
	self.env = env
	if env == "development" {
		self.logLevel = 100
	}
	return self
}

func (self *Web) DB() *kol.DB {
	return self.db
}

func (self *Web) SetSecret(secret string) *Web {
	self.sessionStore = sessions.NewCookieStore([]byte(secret))
	return self
}

func (self *Web) SetAppcache(appcache bool) *Web {
	self.appcache = appcache
	return self
}

func (self *Web) renderText(w http.ResponseWriter, r *http.Request, templates *template.Template, template string, data interface{}) {
	if err := templates.ExecuteTemplate(w, template, data); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
}

func (self *Web) Logf(level int, format string, args ...interface{}) {
	if level <= self.logLevel {
		log.Printf(format, args...)
	}
}

func (self *Web) Fatalf(format string, args ...interface{}) {
	self.Logf(Fatal, "\033[1;31mFATAL\t"+format+"\033[0m", args...)
}

func (self *Web) Errorf(format string, args ...interface{}) {
	self.Logf(Error, "\033[31mERROR\t"+format+"\033[0m", args...)
}

func (self *Web) Infof(format string, args ...interface{}) {
	self.Logf(Info, "INFO\t"+format, args...)
}

func (self *Web) Debugf(format string, args ...interface{}) {
	self.Logf(Debug, "\033[32mDEBUG\t"+format+"\033[0m", args...)
}

func (self *Web) Tracef(format string, args ...interface{}) {
	self.Logf(Trace, "\033[1;32mTRACE\t"+format+"\033[0m", args...)
}

func (self *Web) render_Templates(data RequestData) {
	fmt.Fprintln(data.response, "(function() {")
	fmt.Fprintln(data.response, "  var n;")
	var buf *bytes.Buffer
	var rendered string
	for _, templ := range self._Templates.Templates() {
		fmt.Fprintf(data.response, "  n = $('<script type=\"text/template\" id=\"%v_underscore\"></script>');\n", strings.Split(templ.Name(), ".")[0])
		fmt.Fprintf(data.response, "  n.text('")
		buf = new(bytes.Buffer)
		templ.Execute(buf, data)
		rendered = string(buf.Bytes())
		rendered = spaceRegexp.ReplaceAllString(rendered, " ")
		rendered = strings.Replace(rendered, "\\", "\\\\", -1)
		rendered = strings.Replace(rendered, "'", "\\'", -1)
		fmt.Fprint(data.response, rendered)
		fmt.Fprintln(data.response, "');")
		fmt.Fprintln(data.response, "  $('head').append(n);")
	}
	fmt.Fprintln(data.response, "})();")
}

func (self *Web) HandleStatic(router *mux.Router, dir string) {
	static, err := os.Open(dir)
	if err != nil {
		panic(err)
	}
	children, err := static.Readdirnames(-1)
	if err != nil {
		panic(err)
	}
	for _, fil := range children {
		cpy := fil
		router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
			return strings.HasSuffix(r.URL.Path, cpy)
		}).HandlerFunc(self.Logger(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=864000")
			if strings.HasSuffix(r.URL.Path, ".css") {
				w.Header().Set("Content-Type", "text/css; charset=UTF-8")
			} else if strings.HasSuffix(r.URL.Path, ".js") {
				w.Header().Set("Content-Type", "application/javascript; charset=UTF-8")
			} else if strings.HasSuffix(r.URL.Path, ".png") {
				w.Header().Set("Content-Type", "image/png")
			} else if strings.HasSuffix(r.URL.Path, ".gif") {
				w.Header().Set("Content-Type", "image/gif")
			} else if strings.HasSuffix(r.URL.Path, ".woff") {
				w.Header().Set("Content-Type", "application/font-woff")
			} else {
				w.Header().Set("Content-Type", "application/octet-stream")
			}
			if in, err := os.Open(filepath.Join("static", cpy)); err != nil {
				self.Errorf("%v", err)
				w.WriteHeader(500)
				fmt.Fprintln(w, err)
			} else {
				defer in.Close()
				if _, err := io.Copy(w, in); err != nil {
					self.Errorf("%v", err)
					w.WriteHeader(500)
					fmt.Println(w, err)
				}
			}
		}))
	}
}
