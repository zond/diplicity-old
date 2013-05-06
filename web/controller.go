package web

import (
	"appengine"
	"net/http"
)

func index(w http.ResponseWriter, r *http.Request) {
	data := getRequestData(w, r)
	renderHtml(w, r, htmlTemplates, "index.html", data)
}

func allJs(w http.ResponseWriter, r *http.Request) {
	if !appengine.IsDevAppServer() {
		w.Header().Set("Cache-Control", "public, max-age=864000")
	}
	data := getRequestData(w, r)
	w.Header().Set("Content-Type", "application/javascript; charset=UTF-8")
	renderText(w, r, jsTemplates, "jquery-2.0.0.min.js", data)
	renderText(w, r, jsTemplates, "jquery.mobile-1.3.1.min.js", data)
	renderText(w, r, jsTemplates, "modernizr.js", data)
	renderText(w, r, jsTemplates, "jquery.panzoom.min.js", data)
	renderText(w, r, jsTemplates, "app.js", data)
}

func allCss(w http.ResponseWriter, r *http.Request) {
	if !appengine.IsDevAppServer() {
		w.Header().Set("Cache-Control", "public, max-age=864000")
	}
	data := getRequestData(w, r)
	w.Header().Set("Content-Type", "text/css; charset=UTF-8")
	renderText(w, r, cssTemplates, "jquery.mobile-1.3.1.min.css", data)
	renderText(w, r, cssTemplates, "common.css", data)
}
