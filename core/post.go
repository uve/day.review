package core

import (
	"appengine"
	"net/http"
)

func post(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)

	err := createPost(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	c.Debugf("result: post created")

	return nil
}

func createPost(c appengine.Context) error {
	err := login(c)
	if err != nil {
		return err
	}

	c.Debugf("result: logged successfully")

	mediaID, err := uploadPhoto(c)
	if err != nil {
		return err
	}

	c.Debugf("result: photo uploaded: ", mediaID)

	return nil
}
