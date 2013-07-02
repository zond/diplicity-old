package web

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/openid"
	"github.com/zond/diplicity/user"
	"github.com/zond/kcwraps/kol"
	"io"
	"log"
	"net/http"
	"net/url"
)

func WS(ws *websocket.Conn) {
	session, _ := sessionStore.Get(ws.Request(), SessionName)
	log.Printf("%v\t%v\t%v <-", ws.Request().URL, ws.Request().RemoteAddr, session.Values[SessionEmail])
	var message common.JsonMessage
	var err error
	for {
		if err = websocket.JSON.Receive(ws, &message); err == nil {
			switch message.Type {
			case common.SubscribeType:
				switch message.Subscribe.URI {
				case "/games/open":
					game.SubscribeOpen(common.NewWSSubscription(ws, message.Subscribe.URI))
				case "/user":
					user.SubscribeEmail(common.NewWSSubscription(ws, message.Subscribe.URI), session.Values[SessionEmail])
				default:
					log.Printf("Unrecognized URI: %+v", message.Subscribe.URI)
				}
			case common.UnsubscribeType:
				common.Unsubscribe(ws, message.Subscribe.URI)
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

func Openid(w http.ResponseWriter, r *http.Request) {
	data := GetRequestData(w, r)
	redirect, email, ok := openid.VerifyAuth(r)
	if ok {
		data.Session.Values[SessionEmail] = email
		user := &user.User{
			Id:    []byte(email),
			Email: email,
		}
		if err := common.DB.Get(user); err == kol.NotFound {
			if err = common.DB.Set(user); err != nil {
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

func Logout(w http.ResponseWriter, r *http.Request) {
	data := GetRequestData(w, r)
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

func Login(w http.ResponseWriter, r *http.Request) {
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

func Index(w http.ResponseWriter, r *http.Request) {
	data := GetRequestData(w, r)
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	renderText(w, r, htmlTemplates, "index.html", data)
}

func AppCache(w http.ResponseWriter, r *http.Request) {
	data := GetRequestData(w, r)
	w.Header().Set("Content-Type", "AddType text/cache-manifest .appcache; charset=UTF-8")
	renderText(w, r, textTemplates, "diplicity.appcache", data)
}

func AllJs(w http.ResponseWriter, r *http.Request) {
	data := GetRequestData(w, r)
	w.Header().Set("Cache-Control", "public, max-age=864000")
	w.Header().Set("Content-Type", "application/javascript; charset=UTF-8")
	renderText(w, r, jsTemplates, "jquery-2.0.0.min.js", data)
	renderText(w, r, jsTemplates, "pre_jquery_mobile.js", data)
	renderText(w, r, jsTemplates, "jquery.mobile-1.3.1.min.js", data)
	renderText(w, r, jsTemplates, "jquery.hammer.min.js", data)
	renderText(w, r, jsTemplates, "underscore-min.js", data)
	renderText(w, r, jsTemplates, "backbone-min.js", data)
	renderText(w, r, jsTemplates, "util.js", data)
	renderText(w, r, jsTemplates, "app.js", data)
	render_Templates(data)
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

func AllCss(w http.ResponseWriter, r *http.Request) {
	data := GetRequestData(w, r)
	w.Header().Set("Cache-Control", "public, max-age=864000")
	w.Header().Set("Content-Type", "text/css; charset=UTF-8")
	renderText(w, r, cssTemplates, "jquery.mobile-1.3.1.min.css", data)
	renderText(w, r, cssTemplates, "common.css", data)
}
