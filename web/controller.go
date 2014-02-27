package web

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/user"
	"github.com/zond/gopenid"
	"github.com/zond/wsubs/gosubs"
)

type AdminGameState struct {
	Game    *game.Game
	Phases  game.Phases
	Members []game.MemberState
}

func (self *Web) AdminGetGame(c *Context) (err error) {
	gameId, err := base64.URLEncoding.DecodeString(c.Vars()["game_id"])
	if err != nil {
		return
	}
	g := &game.Game{Id: gameId}
	if err = c.DB().Get(g); err != nil {
		return
	}
	members, err := g.Members(c.DB())
	if err != nil {
		return
	}
	memberStates, err := members.ToStates(c.DB(), g, "")
	if err != nil {
		return
	}
	phases, err := g.Phases(c.DB())
	if err != nil {
		return
	}
	sort.Sort(phases)
	return c.RenderJSON(AdminGameState{
		Game:    g,
		Phases:  phases,
		Members: memberStates,
	})
}

func (self *Web) Openid(c *Context) (err error) {
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
	return
}

func (self *Web) Token(c *Context) (err error) {
	if emailIf, found := c.session.Values[SessionEmail]; found {
		token := &gosubs.Token{
			Principal: fmt.Sprint(emailIf),
			Timeout:   time.Now().Add(time.Second * 10),
		}
		if err = token.Encode(); err != nil {
			return
		}
		err = c.RenderJSON(token)
	} else {
		err = c.RenderJSON(gosubs.Token{})
	}
	return
}

func (self *Web) Logout(c *Context) (err error) {
	delete(c.session.Values, SessionEmail)
	c.Close()
	redirect := c.Req().FormValue("return_to")
	if redirect == "" {
		redirect = fmt.Sprintf("http://%v/", c.Req().Host)
	}
	c.Resp().Header().Set("Location", redirect)
	c.Resp().WriteHeader(302)
	fmt.Fprintln(c.Resp(), redirect)
	return
}

func (self *Web) Login(c *Context) (err error) {
	redirect := c.Req().FormValue("return_to")
	if redirect == "" {
		redirect = fmt.Sprintf("http://%v/", c.Req().Host)
	}
	redirectUrl, err := url.Parse(redirect)
	if err != nil {
		return
	}
	url := gopenid.GetAuthURL(c.Req(), redirectUrl)
	c.Resp().Header().Set("Location", url.String())
	c.Resp().WriteHeader(302)
	fmt.Fprintln(c.Resp(), url.String())
	return
}

func (self *Web) Index(c *Context) (err error) {
	c.SetContentType("text/html; charset=UTF-8", false)
	c.RenderText(self.htmlTemplates, "index.html")
	return
}

func (self *Web) AppCache(c *Context) (err error) {
	if self.appcache {
		c.Resp().Header().Set("Content-Type", "AddType text/cache-manifest .appcache; charset=UTF-8")
		c.RenderText(self.textTemplates, "diplicity.appcache")
	} else {
		c.Resp().WriteHeader(404)
	}
	return
}

func (self *Web) AllJs(c *Context) (err error) {
	c.SetContentType("application/javascript; charset=UTF-8", true)
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
		if err = templ.Execute(c.Resp(), c); err != nil {
			return
		}
	}
	for _, templ := range self.jsCollectionTemplates.Templates() {
		if err = templ.Execute(c.Resp(), c); err != nil {
			return
		}
	}
	for _, templ := range self.jsViewTemplates.Templates() {
		if err = templ.Execute(c.Resp(), c); err != nil {
			return
		}
	}
	c.RenderJS("app.js")
	return
}

func (self *Web) AllCss(c *Context) (err error) {
	c.Resp().Header().Set("Cache-Control", "public, max-age=864000")
	c.Resp().Header().Set("Content-Type", "text/css; charset=UTF-8")
	c.RenderText(self.cssTemplates, "bootstrap.min.css")
	c.RenderText(self.cssTemplates, "bootstrap-theme.min.css")
	c.RenderText(self.cssTemplates, "bootstrap-multiselect.css")
	c.RenderText(self.cssTemplates, "slider.css")
	c.RenderText(self.cssTemplates, "common.css")
	return
}
