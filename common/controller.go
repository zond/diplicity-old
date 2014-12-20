package common

import (
	"io"
	"os"
	"path/filepath"
	"text/template"
)

var temps = template.Must(template.ParseGlob("templates/*"))

func (self *Web) Index(c *HTTPContext) (err error) {
	c.SetContentType("text/html; charset=UTF-8")
	f, err := os.Open(filepath.Join("static", "index.html"))
	if err != nil {
		return
	}
	defer f.Close()
	if _, err = io.Copy(c.Resp(), f); err != nil {
		return
	}
	return
}

func (self *Web) Go2JS(c *HTTPContext) (err error) {
	c.SetContentType("application/javascript; charset=UTF-8")
	if err = temps.ExecuteTemplate(c.Resp(), "go.js", self); err != nil {
		return
	}
	return
}
