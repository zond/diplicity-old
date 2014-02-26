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
	c.RenderText(self.htmlTemplates, "index.html")
}

func (self *Web) AppCache(c *Context) {
	if self.appcache {
		c.Resp().Header().Set("Content-Type", "AddType text/cache-manifest .appcache; charset=UTF-8")
		c.RenderText(self.textTemplates, "diplicity.appcache")
	} else {
		c.Resp().WriteHeader(404)
	}
}

func (self *Web) AllJs(c *Context) {
	common.SetContentType(c.Resp(), "application/javascript; charset=UTF-8", true)
	c.RenderJS("jquery-2.0.3.min.js")
	c.RenderJS("jquery.timeago.js")
	c.RenderJS("jquery.hammer.min.js")
	c.RenderJS("underscore-min.js")
	c.RenderJS("backbone-min.js")
	c.RenderJS("bootstrap.min.js")
	c.RenderJS("bootstrap-multiselect.js")
	c.RenderJS("log.js")
	c.RenderJS("util.js")
	c.RenderJS("panzoom.js")
	c.RenderJS("cache.js")
	c.RenderJS("jsock.js")
	c.RenderJS("wsbackbone.js")
	c.RenderJS("baseView.js")
	c.RenderJS("dippyMap.js")
	self.render_Templates(c)
	for _, templ := range self.jsModelTemplates.Templates() {
		if err := templ.Execute(c.Resp(), c); err != nil {
			panic(err)
		}
	}
	for _, templ := range self.jsCollectionTemplates.Templates() {
		if err := templ.Execute(c.Resp(), c); err != nil {
			panic(err)
		}
	}
	for _, templ := range self.jsViewTemplates.Templates() {
		if err := templ.Execute(c.Resp(), c); err != nil {
			panic(err)
		}
	}
	c.RenderJS("app.js")
}

func (self *Web) AllCss(c *Context) {
	c.Resp().Header().Set("Cache-Control", "public, max-age=864000")
	c.Resp().Header().Set("Content-Type", "text/css; charset=UTF-8")
	c.RenderText(self.cssTemplates, "bootstrap.min.css")
	c.RenderText(self.cssTemplates, "bootstrap-theme.min.css")
	c.RenderText(self.cssTemplates, "bootstrap-multiselect.css")
	c.RenderText(self.cssTemplates, "slider.css")
	c.RenderText(self.cssTemplates, "common.css")
}
