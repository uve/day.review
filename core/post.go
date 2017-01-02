package core

import (
	"appengine"
	"net/http"
)

func postHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	c := appengine.NewContext(r)

	instagramImage := &InstagramImage{
		DisplaySrc: "https://scontent.cdninstagram.com/t51.2885-15/e35/15624166_231201313993189_2257556495791030272_n.jpg?ig_cache_key=MTQxNjg1MDM5MjI2ODc5MTQzNA%3D%3D.2",
		AvatarSrc:  "https://scontent.cdninstagram.com/t51.2885-19/s150x150/14719833_310540259320655_1605122788543168512_a.jpg",
		Caption:    "therock.therock",
		Body:       "Repost from @therock.therock text..",
	}

	err = instagramImage.processImage(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = instagramImage.publish(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Debugf("result: post created")
	/*
		err = writeImage(w, instagramImage.Photo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	*/
}

func (instagramImage *InstagramImage) publish(c appengine.Context) error {
	var err error
	session := newSession()

	err = session.login(c)
	if err != nil {
		return err
	}

	c.Debugf("result: logged successfully")

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
