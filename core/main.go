package core

import (
	// "fmt"
	"net/http"
	//"github.com/gedex/go-instagram/instagram"
	//"appengine"
	//"appengine/urlfetch"
)

func init() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {

	err := parser(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	/*
	   c := appengine.NewContext(r)
	   client := urlfetch.Client(c)
	   instagramClient := instagram.NewClient(client)

	   //instagramClient.ClientID = "8794727f2762463e88f6c19fda007b0e"
	   //instagramClient.ClientSecret = "300e4e786ecd4e608070d08c1dc2e17a"
	   instagramClient.AccessToken = "4306170754.8794727.78e3bf8b30dd4e6b9082dde0e3834c1f"

	   // Gets user's feed.

	   opt := &instagram.Parameters{Count: 3}
	   media, next, err := instagramClient.Users.RecentMedia("25025320", opt)
	   fmt.Fprint(w, "Hello, world!", media, next)

	   if err != nil {
	           http.Error(w, err.Error(), http.StatusInternalServerError)
	           return
	   }
	   fmt.Fprint(w, "Hello, world!", media)*/
}
