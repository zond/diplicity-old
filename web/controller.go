package web

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/openid"
	"github.com/zond/diplicity/user"
	"github.com/zond/kcwraps/subs"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var currentGamePattern = regexp.MustCompile("^/games/current/(.*)$")
var gamePattern = regexp.MustCompile("^/games/(.*)$")

func (self *Web) WS(ws *websocket.Conn) {
	session, err := self.sessionStore.Get(ws.Request(), SessionName)
	if err != nil {
		self.Errorf("%v\t%v\t%v", ws.Request().URL, ws.Request().RemoteAddr, err)
	}

	email := ""
	emailIf, loggedIn := session.Values[SessionEmail]
	if loggedIn {
		email = emailIf.(string)
	}

	self.Infof("%v\t%v\t%v <-", ws.Request().URL, ws.Request().RemoteAddr, session.Values[SessionEmail])

	pack := subs.New(self.db, ws).OnUnsubscribe(func(s *subs.Subscription, reason interface{}) {
		self.Errorf("%v\t%v\t%v\t%v\t%v\t[unsubscribing]", ws.Request().URL, ws.Request().RemoteAddr, emailIf, s.Name(), reason)
	})
	defer func() {
		self.Infof("%v\t%v\t%v -> [unsubscribing all]", ws.Request().URL, ws.Request().RemoteAddr, session.Values[SessionEmail])
		pack.UnsubscribeAll()
	}()

	for {
		var message subs.Message
		if err = websocket.JSON.Receive(ws, &message); err == nil {
			self.Debugf("%v\t%v\t%v\t%v\t%v", ws.Request().URL, ws.Request().RemoteAddr, emailIf, message.Type, message.Object.URI)
			switch message.Type {
			case common.SubscribeType:
				s := pack.New(message.Object.URI)
				switch message.Object.URI {
				case "/games/current":
					if loggedIn {
						self.Errlog(game.SubscribeCurrent(self, s, email))
					}
				case "/games/open":
					if loggedIn {
						self.Errlog(game.SubscribeOpen(self, s, email))
					}
				case "/user":
					if loggedIn {
						self.Errlog(user.SubscribeEmail(self, s, email))
					} else {
						s.Call(&user.User{}, subs.FetchType)
					}
				default:
					if match := gamePattern.FindStringSubmatch(message.Object.URI); match != nil {
						if loggedIn {
							game.SubscribeGame(self, s, match[1], email)
						}
					} else {
						self.Errorf("Unrecognized URI: %+v", message.Object.URI)
					}
				}
			case common.UnsubscribeType:
				pack.Unsubscribe(message.Object.URI)
			case common.CreateType:
				if self.logLevel > Trace {
					self.Tracef("%+v", common.Prettify(message.Object.Data))
				}
				switch message.Object.URI {
				case "/games":
					if loggedIn {
						game.Create(self, common.JSON{message.Object.Data}, email)
					}
				}
			case common.DeleteType:
				if match := currentGamePattern.FindStringSubmatch(message.Object.URI); match != nil {
					if loggedIn {
						game.DeleteMember(self, match[1], email)
					}
				} else {
					self.Errorf("Unrecognized URI to delete: %v", message.Object.URI)
				}
			case common.UpdateType:
				if strings.Index(message.Object.URI, "/games/open") == 0 {
					if loggedIn {
						game.AddMember(self, common.JSON{message.Object.Data}, email)
					}
				}
			default:
				self.Errorf("Unrecognized message Type: %+v", message.Type)
			}
		} else if err == io.EOF {
			break
		} else {
			self.Errorf("%v", err)
		}
	}
}

func (self *Web) Openid(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	redirect, email, ok := openid.VerifyAuth(r)
	if ok {
		data.session.Values[SessionEmail] = email
		user.EnsureUser(self, email)
	} else {
		delete(data.session.Values, SessionEmail)
	}
	data.Close()
	w.Header().Set("Location", redirect.String())
	w.WriteHeader(302)
	fmt.Fprintln(w, redirect.String())
}

func (self *Web) Logout(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	var redirect *url.URL
	r.ParseForm()
	if returnTo := r.Form.Get("return_to"); returnTo == "" {
		redirect = common.MustParseURL("http://" + r.Host + "/")
	} else {
		redirect = common.MustParseURL(returnTo)
	}
	delete(data.session.Values, SessionEmail)
	data.Close()
	w.Header().Set("Location", redirect.String())
	w.WriteHeader(302)
	fmt.Fprintln(w, redirect.String())
}

func (self *Web) Login(w http.ResponseWriter, r *http.Request) {
	var redirect *url.URL
	r.ParseForm()
	if returnTo := r.Form.Get("return_to"); returnTo == "" {
		redirect = common.MustParseURL("http://" + r.Host + "/")
	} else {
		redirect = common.MustParseURL(returnTo)
	}
	url := openid.GetAuthURL(r, redirect)
	w.Header().Set("Location", url.String())
	w.WriteHeader(302)
	fmt.Fprintln(w, url.String())
}

func (self *Web) Index(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	self.renderText(w, r, self.htmlTemplates, "index.html", data)
}

func (self *Web) AppCache(w http.ResponseWriter, r *http.Request) {
	if self.appcache {
		data := self.GetRequestData(w, r)
		w.Header().Set("Content-Type", "AddType text/cache-manifest .appcache; charset=UTF-8")
		self.renderText(w, r, self.textTemplates, "diplicity.appcache", data)
	} else {
		w.WriteHeader(404)
	}
}

func (self *Web) AllJs(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	w.Header().Set("Cache-Control", "public, max-age=864000")
	w.Header().Set("Content-Type", "application/javascript; charset=UTF-8")
	self.renderText(w, r, self.jsTemplates, "jquery-2.0.3.min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "jquery.hammer.min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "underscore-min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "backbone-min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "bootstrap.min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "util.js", data)
	fmt.Fprintln(w, ";")
	self.render_Templates(data)
	fmt.Fprintln(w, ";")
	for _, templ := range self.jsModelTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
		fmt.Fprintln(w, ";")
	}
	for _, templ := range self.jsCollectionTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
		fmt.Fprintln(w, ";")
	}
	for _, templ := range self.jsViewTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
		fmt.Fprintln(w, ";")
	}
	self.renderText(w, r, self.jsTemplates, "app.js", data)
	fmt.Fprintln(w, ";")
}

func (self *Web) AllCss(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	w.Header().Set("Cache-Control", "public, max-age=864000")
	w.Header().Set("Content-Type", "text/css; charset=UTF-8")
	self.renderText(w, r, self.cssTemplates, "bootstrap.min.css", data)
	self.renderText(w, r, self.cssTemplates, "bootstrap-theme.min.css", data)
	self.renderText(w, r, self.cssTemplates, "common.css", data)
}
