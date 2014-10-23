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
