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
	session := newSession()

	err := session.login(c)
	if err != nil {
		return err
	}

	c.Debugf("result: logged successfully")

	mediaID, err := session.uploadPhoto(c)
	if err != nil {
		return err
	}

	c.Debugf("result: photo uploaded: ", mediaID)

	err = session.configurePhoto(c, "test5")
	if err != nil {
		return err
	}

	c.Debugf("Done. Photo configured")

	return nil
}
