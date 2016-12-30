package core

import (
	"appengine"
	"appengine/urlfetch"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	//"strings"
	"net/url"
	"os"
	"time"
)

type loginParamsStruct struct {
	DeviceID         string `json:"device_id"`
	UUID             string `json:"_uuid"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	CSRFToken        string `json:"_csrftoken"`
	LoginAttempCount int    `json:"login_attempt_count"`
}

type configureParamsStruct struct {
	DeviceID        string `json:"device_id"`
	UUID            string `json:"_uuid"`
	CSRFToken       string `json:"_csrftoken"`
	MediaID         string `json:"media_id"`
	Caption         string `json:"caption"`
	DeviceTimestamp int64  `json:"device_timestamp"`
	SourceType      string `json:"source_type"`
	FilterType      string `json:"filter_type"`
	Extra           string `json:"extra"`
	ContentType     string `json:"Content-Type"`
}

type loginStatusStruct struct {
	Status    string `json:"status"`
	ErrorType string `json:"error_type"`
	Message   string `json:"message"`
}

type uploadPhotoStatusStruct struct {
	Status  string `json:"status"`
	MediaID string `json:"upload_id"`
	Message string `json:"message"`
}

var (
	_Versions = []string{
		"GT-N7000",
		"SM-N9000",
		"GT-I9220",
		"GT-I9100",
	}
	_Resolutions = []string{
		"720x1280",
		"320x480",
		"480x800",
		"1024x768",
		"1280x720",
		"768x1024",
		"480x320",
	}
	_Dpis        = []string{"120", "160", "320", "240"}
	_InstVersion = "9.0.1"

	_EndPointURL = "https://i.instagram.com/api/v1"
	_DeviceID    = "android-8c69989e4115c78b"
	_UserName    = "day.review"
	_Password    = "vavilon"
	_ContentType = "application/x-www-form-urlencoded; charset=UTF-8"
)

type Session struct {
	Cookies   []*http.Cookie
	DeviceID  string
	UUID      string
	UserAgent string
	MediaID   string
}

func newSession() *Session {
	session := new(Session)
	session.DeviceID = _DeviceID
	session.UUID = generateUUID()
	session.UserAgent = generateUserAgent()

	return session
}

func generateUserAgent() string {
	var ver = _Versions[rand.Intn(len(_Versions))]
	var res = _Resolutions[rand.Intn(len(_Resolutions))]
	var dpi = _Dpis[rand.Intn(len(_Dpis))]

	var randomVersion = rand.Intn(1) + 10
	var randomVal1 = rand.Intn(3) + 1
	var randomVal2 = rand.Intn(3) + 3
	var randomVal3 = rand.Intn(6)

	var andVersion = fmt.Sprintf("%d/%d.%d.%d", randomVersion, randomVal1, randomVal2, randomVal3)
	var result = fmt.Sprintf("Instagram %s Android (%s; %s; %s; samsung; %s; %s; smdkc210; en_US)", _InstVersion, andVersion, dpi, res, ver, ver)

	return result
}

func generateSignature(message []byte) string {
	key := []byte("96724bcbd4fb3e608074e185f2d4f119156fcca061692a4a1db1c7bf142d3e22")
	sig := hmac.New(sha256.New, key)
	sig.Write(message)

	return hex.EncodeToString(sig.Sum(nil))
}

func generateUUID() (uuid string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	uuid = fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return
}

func (session *Session) login(c appengine.Context) error {
	var loginParams = loginParamsStruct{
		DeviceID:         session.DeviceID,
		UUID:             session.UUID,
		Username:         _UserName,
		Password:         _Password,
		CSRFToken:        "missing",
		LoginAttempCount: 0,
	}

	data, _ := json.Marshal(loginParams)

	var sig = generateSignature(data)
	var payload = fmt.Sprintf("signed_body=%s.%s&ig_sig_key_version=4", sig, string(data))

	client := urlfetch.Client(c)

	urlStr := _EndPointURL + "/accounts/login/"
	req, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(payload))
	req.Header.Add("Content-Type", _ContentType)
	req.Header.Add("User-Agent", session.UserAgent)
	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	c.Debugf(string(body))

	var loginStatus *loginStatusStruct
	err = json.Unmarshal(body, &loginStatus)
	if err != nil {
		return err
	}

	if loginStatus.Status != "ok" {
		return errors.New(loginStatus.Message)
	}

	session.Cookies = resp.Cookies()

	return nil
}

func (session *Session) uploadPhoto(c appengine.Context) (string, error) {
	var now = time.Now().Unix()

	var photoName = "pending_media_1483124562.57.jpg"
	var filename = fmt.Sprintf("%v.2", now)
	var uploadID = fmt.Sprintf("%d", now)

	c.Debugf("DATA")
	c.Debugf(filename)
	c.Debugf(uploadID)

	client := urlfetch.Client(c)
	urlStr := _EndPointURL + "/upload/photo/"

	//////// READ FILE ///////

	file, err := os.Open("static/" + photoName)
	if err != nil {
		return "", err
	}
	fileContents, err := ioutil.ReadAll(file)

	c.Debugf(fmt.Sprintf("Size of file: %v", len(fileContents)))

	uploadBody := new(bytes.Buffer)
	writer := multipart.NewWriter(uploadBody)
	part, err := writer.CreateFormFile("photo", photoName)
	if err != nil {
		return "", err
	}
	part.Write(fileContents)

	//////// READ FILE ///////

	extraParams := map[string]string{
		"_csrftoken":        "missing",
		"upload_id":         uploadID,
		"device_id":         session.DeviceID,
		"_uuid":             session.UUID,
		"image_compression": `{"lib_name":"jt","lib_version":"1.3.0","quality":"70"}`,
		"filename":          filename,
	}

	for key, val := range extraParams {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest("POST", urlStr, uploadBody)
	req.Header.Add("Content-Type", _ContentType)
	req.Header.Add("User-Agent", session.UserAgent)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	for name, cookie := range session.Cookies {
		c.Debugf(fmt.Sprintf("Header %v: %v", name, cookie))
		req.AddCookie(cookie)
	}

	c.Debugf("FormData: %v", writer.FormDataContentType())

	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	c.Debugf("Upload response: %v", string(body))

	var uploadPhotoStatus *uploadPhotoStatusStruct
	err = json.Unmarshal(body, &uploadPhotoStatus)
	if err != nil {
		return "", err
	}

	if uploadPhotoStatus.MediaID == "" {
		return "", errors.New(uploadPhotoStatus.Message)
	}

	session.MediaID = uploadPhotoStatus.MediaID
	return uploadPhotoStatus.MediaID, nil
}

func (session *Session) configurePhoto(c appengine.Context, caption string) error {
	var configureParams = configureParamsStruct{
		DeviceID:        session.DeviceID,
		UUID:            session.UUID,
		CSRFToken:       "missing",
		MediaID:         session.MediaID,
		Caption:         caption,
		DeviceTimestamp: time.Now().Unix(),
		SourceType:      "5",
		FilterType:      "0",
		Extra:           "{}",
		ContentType:     _ContentType,
	}

	data, _ := json.Marshal(configureParams)

	var sig = generateSignature(data)
	var payload = fmt.Sprintf("signed_body=%s.%s&ig_sig_key_version=4", sig, url.QueryEscape(string(data)))

	c.Debugf(fmt.Sprintf("payload", payload))

	client := urlfetch.Client(c)

	urlStr := _EndPointURL + "/media/configure/"
	req, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(payload))
	req.Header.Add("Content-Type", _ContentType)
	req.Header.Add("User-Agent", session.UserAgent)

	for name, cookie := range session.Cookies {
		c.Debugf(fmt.Sprintf("Header %v: %v", name, cookie))
		req.AddCookie(cookie)
	}

	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	c.Debugf(string(body))

	var loginStatus *loginStatusStruct
	err = json.Unmarshal(body, &loginStatus)
	if err != nil {
		return err
	}

	if loginStatus.Status != "ok" {
		return errors.New(loginStatus.Message)
	}

	return nil
}
