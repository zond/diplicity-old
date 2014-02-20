package web

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	"github.com/zond/gopenid"
	"github.com/zond/kcwraps/subs"
)

func (self *Web) Openid(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	redirect, email, ok := gopenid.VerifyAuth(r)
	if ok {
		data.session.Values[SessionEmail] = email
		user.EnsureUser(self.DB(), email)
	} else {
		delete(data.session.Values, SessionEmail)
	}
	data.Close()
	w.Header().Set("Location", redirect.String())
	w.WriteHeader(302)
	fmt.Fprintln(w, redirect.String())
}

func (self *Web) Token(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	if emailIf, found := data.session.Values[SessionEmail]; found {
		token := &subs.Token{
			Principal: fmt.Sprint(emailIf),
			Timeout:   time.Now().Add(time.Second * 10),
		}
		if err := token.Encode(); err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err)
			return
		}
		common.RenderJSON(w, token)
	} else {
		common.RenderJSON(w, subs.Token{})
	}
}

func (self *Web) Logout(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	delete(data.session.Values, SessionEmail)
	data.Close()
	redirect := r.FormValue("return_to")
	if redirect == "" {
		redirect = fmt.Sprintf("http://%v/", r.Host)
	}
	w.Header().Set("Location", redirect)
	w.WriteHeader(302)
	fmt.Fprintln(w, redirect)
}

func (self *Web) Login(w http.ResponseWriter, r *http.Request) {
	redirect := r.FormValue("return_to")
	if redirect == "" {
		redirect = fmt.Sprintf("http://%v/", r.Host)
	}
	redirectUrl, err := url.Parse(redirect)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
		return
	}
	url := gopenid.GetAuthURL(r, redirectUrl)
	w.Header().Set("Location", url.String())
	w.WriteHeader(302)
	fmt.Fprintln(w, url.String())
}

func (self *Web) Index(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	common.SetContentType(w, "text/html; charset=UTF-8", false)
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
	common.SetContentType(w, "application/javascript; charset=UTF-8", true)
	self.renderText(w, r, self.jsTemplates, "jquery-2.0.3.min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "jquery.timeago.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "jquery.hammer.min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "underscore-min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "backbone-min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "bootstrap.min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "bootstrap-multiselect.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "log.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "util.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "panzoom.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "cache.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "wsBackbone.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "baseView.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "dippyMap.js", data)
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
	self.renderText(w, r, self.cssTemplates, "bootstrap-multiselect.css", data)
	self.renderText(w, r, self.cssTemplates, "slider.css", data)
	self.renderText(w, r, self.cssTemplates, "common.css", data)
}
