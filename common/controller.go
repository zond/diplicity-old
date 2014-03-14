package common

func (self *Web) Index(c *HTTPContext) (err error) {
	c.SetContentType("text/html; charset=UTF-8", false)
	c.RenderText(self.htmlTemplates, "index.html")
	return
}

func (self *Web) AppCache(c *HTTPContext) (err error) {
	if self.appcache {
		c.Resp().Header().Set("Content-Type", "AddType text/cache-manifest .appcache; charset=UTF-8")
		c.RenderText(self.textTemplates, "diplicity.appcache")
	} else {
		c.Resp().WriteHeader(404)
	}
	return
}

func (self *Web) AllJs(c *HTTPContext) (err error) {
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

func (self *Web) AllCss(c *HTTPContext) (err error) {
	c.Resp().Header().Set("Cache-Control", "public, max-age=864000")
	c.Resp().Header().Set("Content-Type", "text/css; charset=UTF-8")
	c.RenderText(self.cssTemplates, "bootstrap.min.css")
	c.RenderText(self.cssTemplates, "bootstrap-theme.min.css")
	c.RenderText(self.cssTemplates, "bootstrap-multiselect.css")
	c.RenderText(self.cssTemplates, "slider.css")
	c.RenderText(self.cssTemplates, "common.css")
	return
}
