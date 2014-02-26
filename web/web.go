package web

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/wsubs/gosubs"
)

const (
	SessionEmail = "email"
	SessionName  = "diplicity_session"
)

const (
	Fatal = iota
	Error
	Info
	Debug
	Trace
)

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
		sessionStore:          sessions.NewCookieStore([]byte(gosubs.Secret)),
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

func (self *Web) SetAppcache(appcache bool) *Web {
	self.appcache = appcache
	return self
}

func (self *Web) renderText(c *Context, templates *template.Template, template string) {
	if err := templates.ExecuteTemplate(c.Resp(), template, c); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
}

func (self *Web) Handle(r *mux.Route, f func(c *Context)) {
	r.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			request:        r,
			start:          time.Now(),
			status:         200,
			web:            self,
		}
		var i int64
		defer func() {
			atomic.StoreInt64(&i, 1)
			lw.log(recover())
		}()
		go func() {
			time.Sleep(time.Second)
			if atomic.CompareAndSwapInt64(&i, 0, 1) {
				lw.inc()
			}
		}()
		f(self.GetContext(w, r))
	})
}

func (self *Web) Logf(level int, format string, args ...interface{}) {
	if level <= self.logLevel {
		log.Printf(format, args...)
	}
}

func (self *Web) Errlog(err error) {
	if err != nil {
		self.Errorf("%v", err)
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

func (self *Web) render_Templates(data *Context) {
	fmt.Fprintln(data.response, "(function() {")
	fmt.Fprintln(data.response, "  var n;")
	var buf *bytes.Buffer
	var rendered string
	for _, templ := range self._Templates.Templates() {
		fmt.Fprintf(data.response, "  n = $('<script type=\"text/template\" id=\"%v_underscore\"></script>');\n", strings.Split(templ.Name(), ".")[0])
		fmt.Fprintf(data.response, "  n.text('")
		buf = new(bytes.Buffer)
		if err := templ.Execute(buf, data); err != nil {
			panic(err)
		}
		rendered = string(buf.Bytes())
		rendered = strings.Replace(rendered, "\\", "\\\\", -1)
		rendered = strings.Replace(rendered, "'", "\\'", -1)
		rendered = strings.Replace(rendered, "\n", "\\n", -1)
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
		self.Handle(router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
			return strings.HasSuffix(r.URL.Path, cpy)
		}), func(c *Context) {
			if strings.HasSuffix(c.Req().URL.Path, ".css") {
				common.SetContentType(c.Resp(), "text/css; charset=UTF-8", true)
			} else if strings.HasSuffix(c.Req().URL.Path, ".js") {
				common.SetContentType(c.Resp(), "application/javascript; charset=UTF-8", true)
			} else if strings.HasSuffix(c.Req().URL.Path, ".png") {
				common.SetContentType(c.Resp(), "image/png", true)
			} else if strings.HasSuffix(c.Req().URL.Path, ".gif") {
				common.SetContentType(c.Resp(), "image/gif", true)
			} else if strings.HasSuffix(c.Req().URL.Path, ".c.Resp()off") {
				common.SetContentType(c.Resp(), "application/font-c.Resp()off", true)
			} else if strings.HasSuffix(c.Req().URL.Path, ".ttf") {
				common.SetContentType(c.Resp(), "font/truetype", true)
			} else {
				common.SetContentType(c.Resp(), "application/octet-stream", true)
			}
			if in, err := os.Open(filepath.Join("static", cpy)); err != nil {
				self.Errorf("%v", err)
				c.Resp().WriteHeader(500)
				fmt.Fprintln(c.Resp(), err)
			} else {
				defer in.Close()
				if _, err := io.Copy(c.Resp(), in); err != nil {
					self.Errorf("%v", err)
					c.Resp().WriteHeader(500)
					fmt.Println(c.Resp(), err)
				}
			}
		})
	}
}
