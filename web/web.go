package web

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world")
}

func init() {
	router := mux.NewRouter()
	router.Path("/").HandlerFunc(index)
	http.Handle("/", router)
}
