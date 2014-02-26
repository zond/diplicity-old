package web

import (
	"fmt"
	"net/url"
	"time"

	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	"github.com/zond/gopenid"
	"github.com/zond/wsubs/gosubs"
)

func (self *Web) Openid(c *Context) {
	redirect, email, ok := gopenid.VerifyAuth(c.Req())
	if ok {
		c.session.Values[SessionEmail] = email
		user.EnsureUser(self.DB(), email)
	} else {
		delete(c.session.Values, SessionEmail)
	}
	c.Close()
	c.Resp().Header().Set("Location", redirect.String())
	c.Resp().WriteHeader(302)
	fmt.Fprintln(c.Resp(), redirect.String())
}

func (self *Web) Token(c *Context) {
	if emailIf, found := c.session.Values[SessionEmail]; found {
		token := &gosubs.Token{
			Principal: fmt.Sprint(emailIf),
			Timeout:   time.Now().Add(time.Second * 10),
		}
		if err := token.Encode(); err != nil {
			c.Resp().WriteHeader(500)
			fmt.Fprintln(c.Resp(), err)
			return
		}
		common.RenderJSON(c.Resp(), token)
	} else {
		common.RenderJSON(c.Resp(), gosubs.Token{})
	}
}

func (self *Web) Logout(c *Context) {
	delete(c.session.Values, SessionEmail)
	c.Close()
	redirect := c.Req().FormValue("return_to")
	if redirect == "" {
		redirect = fmt.Sprintf("http://%v/", c.Req().Host)
	}
	c.Resp().Header().Set("Location", redirect)
	c.Resp().WriteHeader(302)
	fmt.Fprintln(c.Resp(), redirect)
}

func (self *Web) Login(c *Context) {
	redirect := c.Req().FormValue("return_to")
	if redirect == "" {
		redirect = fmt.Sprintf("http://%v/", c.Req().Host)
	}
	redirectUrl, err := url.Parse(redirect)
	if err != nil {
		c.Resp().WriteHeader(500)
		fmt.Fprintln(c.Resp(), err)
		return
	}
	url := gopenid.GetAuthURL(c.Req(), redirectUrl)
	c.Resp().Header().Set("Location", url.String())
	c.Resp().WriteHeader(302)
	fmt.Fprintln(c.Resp(), url.String())
}

func (self *Web) Index(c *Context) {
	common.SetContentType(c.Resp(), "text/html; charset=UTF-8", false)
	self.renderText(c, self.htmlTemplates, "index.html")
}

func (self *Web) AppCache(c *Context) {
	if self.appcache {
		c.Resp().Header().Set("Content-Type", "AddType text/cache-manifest .appcache; charset=UTF-8")
		self.renderText(c, self.textTemplates, "diplicity.appcache")
	} else {
		c.Resp().WriteHeader(404)
	}
}

func (self *Web) AllJs(c *Context) {
	common.SetContentType(c.Resp(), "application/javascript; charset=UTF-8", true)
	self.renderText(c, self.jsTemplates, "jquery-2.0.3.min.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "jquery.timeago.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "jquery.hammer.min.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "underscore-min.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "backbone-min.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "bootstrap.min.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "bootstrap-multiselect.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "log.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "util.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "panzoom.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "cache.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "jsock.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "wsbackbone.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "baseView.js")
	fmt.Fprintln(c.Resp(), ";")
	self.renderText(c, self.jsTemplates, "dippyMap.js")
	fmt.Fprintln(c.Resp(), ";")
	self.render_Templates(c)
	fmt.Fprintln(c.Resp(), ";")
	for _, templ := range self.jsModelTemplates.Templates() {
		if err := templ.Execute(c.Resp(), c); err != nil {
			panic(err)
		}
		fmt.Fprintln(c.Resp(), ";")
	}
	for _, templ := range self.jsCollectionTemplates.Templates() {
		if err := templ.Execute(c.Resp(), c); err != nil {
			panic(err)
		}
		fmt.Fprintln(c.Resp(), ";")
	}
	for _, templ := range self.jsViewTemplates.Templates() {
		if err := templ.Execute(c.Resp(), c); err != nil {
			panic(err)
		}
		fmt.Fprintln(c.Resp(), ";")
	}
	self.renderText(c, self.jsTemplates, "app.js")
	fmt.Fprintln(c.Resp(), ";")
}

func (self *Web) AllCss(c *Context) {
	c.Resp().Header().Set("Cache-Control", "public, max-age=864000")
	c.Resp().Header().Set("Content-Type", "text/css; charset=UTF-8")
	self.renderText(c, self.cssTemplates, "bootstrap.min.css")
	self.renderText(c, self.cssTemplates, "bootstrap-theme.min.css")
	self.renderText(c, self.cssTemplates, "bootstrap-multiselect.css")
	self.renderText(c, self.cssTemplates, "slider.css")
	self.renderText(c, self.cssTemplates, "common.css")
}
