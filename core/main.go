package core

import (
	// "fmt"
	"net/http"
	//"github.com/gedex/go-instagram/instagram"
	//"appengine"
	//"appengine/urlfetch"
)

func init() {
	http.HandleFunc("/parser", parserHandler)
	http.HandleFunc("/post", postHandler)
}

func parserHandler(w http.ResponseWriter, r *http.Request) {

	err := parser(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {

	err := post(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
