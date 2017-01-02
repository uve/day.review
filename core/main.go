package core

import (
	"appengine"
	//"fmt"
	"net/http"
)

func init() {
	http.HandleFunc("/parser", parserHandler)
	http.HandleFunc("/image", parserImage)
	http.HandleFunc("/post", postHandler)
}

func parserHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	account, err := parser(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	instagramImage := &InstagramImage{
		DisplaySrc: account.Node.DisplaySrc,
		AvatarSrc:  account.User.ProfilePicURL,
		Caption:    account.User.Username,
		Body:       "Repost from @" + account.User.Username + " " + account.Node.Caption,
	}

	//c.Debugf(fmt.Sprintf("Account: %v", account.Node.Owner))

	err = instagramImage.processImage(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = instagramImage.publish(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	/*
		err = writeImage(w, instagramImage.Photo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	*/

	c.Debugf("result: post created")
}

func parserImage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	instagramImage := &InstagramImage{
		DisplaySrc: "https://scontent.cdninstagram.com/t51.2885-15/e35/15624166_231201313993189_2257556495791030272_n.jpg",
		AvatarSrc:  "https://scontent.cdninstagram.com/t51.2885-19/s150x150/14719833_310540259320655_1605122788543168512_a.jpg",
		Caption:    "therock.therock",
		Body:       "Repost from @therock.therock text..",
	}

	err := instagramImage.processImage(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = writeImage(w, instagramImage.Photo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
