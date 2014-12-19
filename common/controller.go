package common

import (
	"io"
	"os"
	"path/filepath"
)

func (self *Web) Index(c *HTTPContext) (err error) {
	c.SetContentType("text/html; charset=UTF-8", false)
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
