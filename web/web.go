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
	"github.com/zond/diplicity/epoch"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/translation"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/wsubs/gosubs"
)

const (
	SessionEmail = "email"
	SessionName  = "diplicity_session"
	Admin        = "Admin"
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

func (self *Web) Start() (err error) {
	startedAt, err := epoch.Get(self.DB())
	if err != nil {
		return
	}
	self.Debugf("At %v", startedAt)
	startedTime := time.Now()
	var currently time.Duration
	go func() {
		for {
			time.Sleep(time.Minute)
			currently = time.Now().Sub(startedTime) + startedAt
			if err = epoch.Set(self.DB(), currently); err != nil {
				panic(err)
			}
			self.Debugf("At %v", currently)
		}
	}()
	unresolved := game.Phases{}
	if err = self.DB().Query().Where(kol.Equals{"Resolved", false}).All(&unresolved); err != nil {
		return err
	}
	for _, phase := range unresolved {
		phase.Schedule(self)
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

func (self *Web) GetContext(w http.ResponseWriter, r *http.Request) (result *Context) {
	result = &Context{
		response:     w,
		request:      r,
		web:          self,
		translations: translation.GetTranslations(common.GetLanguage(r)),
		vars:         mux.Vars(r),
	}
	result.session, _ = self.sessionStore.Get(r, SessionName)
	return
}

func (self *Web) SetAppcache(appcache bool) *Web {
	self.appcache = appcache
	return self
}

func (self *Web) AdminHandle(r *mux.Route, f func(c *Context) error) {
	self.Handle(r, func(c *Context) (err error) {
		tokenStr := c.Req().FormValue("token")
		if tokenStr == "" {
			err = fmt.Errorf("Missing token")
			return
		}
		token, err := gosubs.DecodeToken(tokenStr)
		if err != nil {
			return
		}
		if token.Principal != Admin {
			err = fmt.Errorf("Not admin")
			return
		}
		return f(c)
	})
}

func (self *Web) Handle(r *mux.Route, f func(c *Context) error) {
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
		if err := f(self.GetContext(w, r)); err != nil {
			lw.WriteHeader(500)
			fmt.Fprintln(lw, err)
			self.Errorf("%v", err)
		}
		return
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
		}), func(c *Context) (err error) {
			if strings.HasSuffix(c.Req().URL.Path, ".css") {
				c.SetContentType("text/css; charset=UTF-8", true)
			} else if strings.HasSuffix(c.Req().URL.Path, ".js") {
				c.SetContentType("application/javascript; charset=UTF-8", true)
			} else if strings.HasSuffix(c.Req().URL.Path, ".png") {
				c.SetContentType("image/png", true)
			} else if strings.HasSuffix(c.Req().URL.Path, ".gif") {
				c.SetContentType("image/gif", true)
			} else if strings.HasSuffix(c.Req().URL.Path, ".c.Resp()off") {
				c.SetContentType("application/font-c.Resp()off", true)
			} else if strings.HasSuffix(c.Req().URL.Path, ".ttf") {
				c.SetContentType("font/truetype", true)
			} else {
				c.SetContentType("application/octet-stream", true)
			}
			in, err := os.Open(filepath.Join("static", cpy))
			if err != nil {
				self.Errorf("%v", err)
				c.Resp().WriteHeader(500)
				fmt.Fprintln(c.Resp(), err)
			} else {
				defer in.Close()
				if _, err = io.Copy(c.Resp(), in); err != nil {
					return
				}
			}
			return
		})
	}
}
