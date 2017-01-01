package core

import (
	"appengine"
	"net/http"
)

func post(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)

	instagramImage := &InstagramImage{
		DisplaySrc: "https://scontent.cdninstagram.com/t51.2885-15/e35/15624166_231201313993189_2257556495791030272_n.jpg?ig_cache_key=MTQxNjg1MDM5MjI2ODc5MTQzNA%3D%3D.2",
		AvatarSrc:  "https://scontent.cdninstagram.com/t51.2885-19/s150x150/14719833_310540259320655_1605122788543168512_a.jpg",
		Caption:    "therock.therock",
		Body:       "Repost from @therock.therock text..",
	}

	err := createPost(c, instagramImage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	c.Debugf("result: post created")

	return nil
}

func createPost(c appengine.Context, instagramImage *InstagramImage) error {
	var err error
	session := newSession()

	err = session.login(c)
	if err != nil {
		return err
	}

	c.Debugf("result: logged successfully")
	/*
		instagramImage.Image, err = instagramImage.load(c, instagramImage.DisplaySrc)
		if err != nil {
			return err
		}
	*/
	mediaID, err := session.uploadPhoto(c, instagramImage.Photo)
	if err != nil {
		return err
	}

	c.Debugf("result: photo uploaded: ", mediaID)

	err = session.configurePhoto(c, instagramImage.Body)
	if err != nil {
		return err
	}

	c.Debugf("Done. Photo configured")

	return nil
}
