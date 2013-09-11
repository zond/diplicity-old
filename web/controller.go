package web

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/openid"
	"github.com/zond/diplicity/user"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
	"io"
	"log"
	"net/http"
	"net/url"
)

func (self *Web) WS(ws *websocket.Conn) {
	session, _ := self.sessionStore.Get(ws.Request(), SessionName)
	log.Printf("%v\t%v\t%v <-", ws.Request().URL, ws.Request().RemoteAddr, session.Values[SessionEmail])

	pack := subs.New(self.db, ws)
	defer pack.UnsubscribeAll()

	email := ""
	emailIf, loggedIn := session.Values[SessionEmail]
	if loggedIn {
		email = emailIf.(string)
	}

	var message common.JsonMessage
	var err error
	for {
		if err = websocket.JSON.Receive(ws, &message); err == nil {
			switch message.Type {
			case common.SubscribeType:
				s := pack.New(message.Subscribe.URI)
				switch message.Subscribe.URI {
				case "/games/current":
					if loggedIn {
						self.SubscribeQuery(game.CurrentSubscription(self.db, s, email))
					}
				case "/games/open":
					if loggedIn {
						self.SubscribeQuery(game.OpenSubscription(self.db, s, email))
					}
				case "/user":
					if loggedIn {
						s.Call(&user.User{}, FetchType)
					} else {
						self.Subscribe(user.EmailSubscription(s, email))
					}
				default:
					log.Printf("Unrecognized URI: %+v", message.Subscribe.URI)
				}
			case common.UnsubscribeType:
				pack.Unsubscribe(message.Subscribe.URI)
			case common.CreateType:
				game.Create(message.Object, session.Values[SessionEmail])
			default:
				log.Printf("Unrecognized message Type: %+v", message.Type)
			}
		} else if err == io.EOF {
			break
		} else {
			log.Println(err)
		}
	}
	log.Printf("%v\t%v\t%v ->", ws.Request().URL, ws.Request().RemoteAddr, session.Values[SessionEmail])
}

func (self *Web) Openid(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	redirect, email, ok := openid.VerifyAuth(r)
	if ok {
		data.Session.Values[SessionEmail] = email
		user := &user.User{
			Id:    []byte(email),
			Email: email,
		}
		if err := self.db.Get(user); err == kol.NotFound {
			if err = self.db.Set(user); err != nil {
				panic(err)
			}
		} else if err != nil {
			panic(err)
		}
	} else {
		delete(data.Session.Values, SessionEmail)
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
	delete(data.Session.Values, SessionEmail)
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
	self.renderText(w, r, htmlTemplates, "index.html", data)
}

func (self *Web) AppCache(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	w.Header().Set("Content-Type", "AddType text/cache-manifest .appcache; charset=UTF-8")
	self.renderText(w, r, textTemplates, "diplicity.appcache", data)
}

func (self *Web) AllJs(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	w.Header().Set("Cache-Control", "public, max-age=864000")
	w.Header().Set("Content-Type", "application/javascript; charset=UTF-8")
	self.renderText(w, r, jsTemplates, "jquery-2.0.0.min.js", data)
	self.renderText(w, r, jsTemplates, "pre_jquery_mobile.js", data)
	self.renderText(w, r, jsTemplates, "jquery.mobile-1.3.1.min.js", data)
	self.renderText(w, r, jsTemplates, "jquery.hammer.min.js", data)
	self.renderText(w, r, jsTemplates, "underscore-min.js", data)
	self.renderText(w, r, jsTemplates, "backbone-min.js", data)
	self.renderText(w, r, jsTemplates, "util.js", data)
	self.renderText(w, r, jsTemplates, "app.js", data)
	self.render_Templates(data)
	for _, templ := range jsModelTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
	}
	for _, templ := range jsCollectionTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
	}
	for _, templ := range jsViewTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
	}
}

func (self *Web) AllCss(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	w.Header().Set("Cache-Control", "public, max-age=864000")
	w.Header().Set("Content-Type", "text/css; charset=UTF-8")
	self.renderText(w, r, cssTemplates, "jquery.mobile-1.3.1.min.css", data)
	self.renderText(w, r, cssTemplates, "common.css", data)
}
