package web

import (
	"github.com/zond/diplicity/common"
	"net/http"
)

func Reload(w http.ResponseWriter, r *http.Request) {
	data := common.GetRequestData(w, r)
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	renderText(w, r, htmlTemplates, "reload.html", data)
}

func Index(w http.ResponseWriter, r *http.Request) {
	data := common.GetRequestData(w, r)
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	renderText(w, r, htmlTemplates, "index.html", data)
}

func AppCache(w http.ResponseWriter, r *http.Request) {
	data := common.GetRequestData(w, r)
	w.Header().Set("Content-Type", "AddType text/cache-manifest .appcache; charset=UTF-8")
	renderText(w, r, textTemplates, "diplicity.appcache", data)
}

func AllJs(w http.ResponseWriter, r *http.Request) {
	data := common.GetRequestData(w, r)
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
	data := common.GetRequestData(w, r)
	w.Header().Set("Cache-Control", "public, max-age=864000")
	w.Header().Set("Content-Type", "text/css; charset=UTF-8")
	renderText(w, r, cssTemplates, "jquery.mobile-1.3.1.min.css", data)
	renderText(w, r, cssTemplates, "common.css", data)
}
