package common

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"path/filepath"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jhillyerd/go.enmime"
	"github.com/zond/diplicity/translation"
	"github.com/zond/gmail"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/templar"
	"github.com/zond/wsubs/gosubs"
)

const (
	SessionEmail = "email"
	SessionName  = "diplicity_session"
	Admin        = "Admin"
	Development  = "development"
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
	gmail                 *gmail.Client
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
	gmailAccount          string
	gmailPassword         string
	smtpAccount           string
	smtpHost              string
	mailHandler           func(c SkinnyContext, msg *enmime.MIMEBody) error
	router                *Router
	secret                string
}

func NewWeb(secret, env, db string) (self *Web, err error) {
	self = &Web{
		secret:       secret,
		appcache:     true,
		env:          env,
		sessionStore: sessions.NewCookieStore([]byte(secret)),
	}
	if self.db, err = kol.New(db); err != nil {
		return
	}
	self.router = newRouter(self)
	self.router.Secret = secret
	if env != Development {
		if secret == DefaultSecret {
			err = fmt.Errorf("Only development env can run with the default secret")
			return
		}
		self.logLevel = Trace
	} else {
		self.logLevel = Debug
		self.router.DevMode = true
	}
	self.router.LogLevel = self.logLevel
	if self.svgTemplates, err = templar.GetMatchingTemplates(env == Development, "svgTemplates", "^templates/svg/[^/]*\\.svg$"); err != nil {
		return
	}
	if self.textTemplates, err = templar.GetMatchingTemplates(env == Development, "textTemplates", "^templates/text/[^/]*$"); err != nil {
		return
	}
	if self.jsModelTemplates, err = templar.GetMatchingTemplates(env == Development, "jsCollectionTemplates", "^templates/js/models/[^/]*\\.js$"); err != nil {
		return
	}
	if self.jsCollectionTemplates, err = templar.GetMatchingTemplates(env == Development, "jsModelTemplates", "^templates/js/collections/[^/]*\\.js$"); err != nil {
		return
	}
	if self.jsTemplates, err = templar.GetMatchingTemplates(env == Development, "jsTemplates", "^templates/js/[^/]*\\.js$"); err != nil {
		return
	}
	if self.cssTemplates, err = templar.GetMatchingTemplates(env == Development, "cssTemplates", "^templates/css/[^/]*\\.css$"); err != nil {
		return
	}
	if self._Templates, err = templar.GetMatchingTemplates(env == Development, "_Templates", "^templates/_/[^/]*\\.html$"); err != nil {
		return
	}
	if self.jsViewTemplates, err = templar.GetMatchingTemplates(env == Development, "jsViewTemplates", "^templates/js/views/[^/]*\\.js$"); err != nil {
		return
	}
	if self.htmlTemplates, err = templar.GetMatchingTemplates(env == Development, "htmlTemplates", "^templates/html/[^/]*\\.html$"); err != nil {
		return
	}
	return
}

func (self *Web) Router() *Router {
	return self.router
}

func (self *Web) Env() string {
	return self.env
}

func (self *Web) IsSubscribing(principal, uri string) bool {
	return self.router.IsSubscribing(principal, uri)
}

func (self *Web) Secret() string {
	return self.secret
}

func (self *Web) MailAddress() string {
	return self.gmailAccount
}

func (self *Web) SetSMTP(host, account string) *Web {
	self.smtpAccount = account
	self.smtpHost = host
	return self
}

func (self *Web) Start() (err error) {
	if self.gmailAccount != "" {
		self.gmail = gmail.New(self.gmailAccount, self.gmailPassword).MailHandler(self.IncomingMail).ErrorHandler(func(e error) {
			self.Fatalf("Mail handler: %v", e)
		})
		if _, err = self.gmail.Start(); err != nil {
			return
		}
		self.Infof("Listening to incoming mail from %#v", self.gmailAccount)
	}
	return
}

func (self *Web) IncomingMail(msg *enmime.MIMEBody) error {
	return self.mailHandler(self.Diet(), msg)
}

func (self *Web) Diet() SkinnyContext {
	return skinnyWeb{
		Web: self,
		db:  self.DB(),
	}
}

func (self *Web) SendMail(fromName, replyTo, subject, message string, recips []string) (err error) {
	body := strings.Join([]string{
		"Content-Type: text/plain; charset=\"utf-8\"",
		fmt.Sprintf("Reply-To: %v", replyTo),
		fmt.Sprintf("From: %v <%v>", fromName, self.smtpAccount),
		fmt.Sprintf("To: %v", strings.Join(recips, ", ")),
		fmt.Sprintf("Subject: %v", subject),
		"",
		message,
	}, "\r\n")
	actualRecips := []string{}
	for _, recip := range recips {
		if match := gmail.AddrReg.FindString(recip); match != "" {
			actualRecips = append(actualRecips, match)
		}
	}
	if self.Env() == Development {
		self.Infof("Would have sent\n%v", body)
	} else {
		self.Infof("Will try to send\n%v", body)
		if err = smtp.SendMail(self.smtpHost, nil, self.smtpAccount, actualRecips, []byte(body)); err != nil {
			self.Errorf("Unable to send\n%v\ndue to %v", body, err)
			return
		}
		self.Infof("Sent\n%v", body)
	}
	return
}

func (self *Web) DB() *kol.DB {
	return self.db
}

func (self *Web) GetContext(w http.ResponseWriter, r *http.Request) (result *HTTPContext) {
	result = &HTTPContext{
		response:     w,
		request:      r,
		web:          self,
		translations: translation.GetTranslations(GetLanguage(r)),
		vars:         mux.Vars(r),
	}
	result.session, _ = self.sessionStore.Get(r, SessionName)
	return
}

func (self *Web) SetGMail(account, password string, handler func(c SkinnyContext, msg *enmime.MIMEBody) error) *Web {
	self.gmailAccount, self.gmailPassword, self.mailHandler = account, password, handler
	return self
}

func (self *Web) SetAppcache(appcache bool) *Web {
	self.appcache = appcache
	return self
}

func (self *Web) DevHandle(r *mux.Route, f func(c *HTTPContext) error) {
	if self.Env() == Development {
		self.Handle(r, func(c *HTTPContext) (err error) {
			if c.Env() == Development {
				return f(c)
			}
			c.Resp().WriteHeader(403)
			return
		})
	}
}

func (self *Web) AdminHandle(r *mux.Route, f func(c *HTTPContext) error) {
	self.Handle(r, func(c *HTTPContext) (err error) {
		tokenStr := c.Req().FormValue("token")
		if tokenStr == "" {
			err = fmt.Errorf("Missing token")
			return
		}
		token, err := gosubs.DecodeToken(self.secret, tokenStr)
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

func (self *Web) Handle(r *mux.Route, f func(c *HTTPContext) error) {
	r.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{
			ResponseWriter: w,
			request:        r,
			start:          time.Now(),
			status:         200,
			web:            self,
		}
		for _, encoding := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
			if strings.TrimSpace(encoding) == "gzip" {
				rw.gzipWriter = gzip.NewWriter(rw.ResponseWriter)
				rw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
				defer rw.Close()
				break
			}
		}
		var i int64
		defer func() {
			atomic.StoreInt64(&i, 1)
			rw.log(recover())
		}()
		go func() {
			time.Sleep(time.Second)
			if atomic.CompareAndSwapInt64(&i, 0, 1) {
				rw.inc()
			}
		}()
		if err := f(self.GetContext(rw, r)); err != nil {
			rw.WriteHeader(500)
			fmt.Fprintln(rw, err)
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

func (self *Web) render_Templates(data *HTTPContext) {
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

func (self *Web) HandleStatic(router *mux.Router, dir string) (err error) {
	children, err := templar.GetMatchingBlobNames(self.env == Development, "^static/.*")
	if err != nil {
		return
	}
	for _, fil := range children {
		cpy := fil
		self.Handle(router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
			return strings.HasSuffix(r.URL.Path, filepath.Base(cpy))
		}), func(c *HTTPContext) (err error) {
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
			in, err := templar.GetBlob(self.env == Development, cpy)
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
	return
}
