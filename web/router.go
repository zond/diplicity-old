package web

import (
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	router := mux.NewRouter()
	router.HandleFunc("/js/{ver}/all", allJs)
	router.HandleFunc("/css/{ver}/all", allCss)
	router.HandleFunc("/", index)
	http.Handle("/", router)
}
