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

type uploadParamsStruct struct {
	DeviceID         string `json:"device_id"`
	UUID             string `json:"_uuid"`
	CSRFToken        string `json:"_csrftoken"`
	UploadID         int64  `json:"upload_id"`
	ImageCompression string `json:"image_compression"`
	Filename         string `json:"filename"`
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
	_uuid        = generateUUID()
	_UserAgent   = generateUserAgent()
)

func login(c appengine.Context) error {
	var loginParams = loginParamsStruct{
		DeviceID:         _DeviceID,
		UUID:             _uuid,
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
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("User-Agent", _UserAgent)
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

func uploadPhoto(c appengine.Context /*, file []byte*/) (string, error) {
	var now = time.Now().Unix()
	var uploadParams = uploadParamsStruct{
		DeviceID:         _DeviceID,
		UUID:             _uuid,
		CSRFToken:        "missing",
		UploadID:         now,
		ImageCompression: `{"lib_name":"jt","lib_version":"1.3.0","quality":"70"}`,
		Filename:         fmt.Sprintf("pending_media_%d.jpg", now),
	}

	//data, _ := json.Marshal(uploadParams)

	client := urlfetch.Client(c)

	urlStr := _EndPointURL + "/upload/photo/"
	//req, _ := http.NewRequest("POST", urlStr, bytes.NewBuffer(data))
	//req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	//req.Header.Add("User-Agent", _UserAgent)
	//resp, _ := client.Do(req)

	fileContents := []byte("mmmm")
	uploadBody := new(bytes.Buffer)
	writer := multipart.NewWriter(uploadBody)
	part, err := writer.CreateFormFile("upload", "static/pic1.jpg")
	if err != nil {
		return "", err
	}

	part.Write(fileContents)

	writer.WriteField("_csrftoken", uploadParams.CSRFToken)
	writer.WriteField("upload_id", string(uploadParams.UploadID))
	writer.WriteField("device_id", uploadParams.DeviceID)
	writer.WriteField("_uuid", uploadParams.UUID)
	writer.WriteField("image_compression", uploadParams.ImageCompression)
	writer.WriteField("filename", uploadParams.Filename)

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest("POST", urlStr, uploadBody)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("User-Agent", _UserAgent)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	c.Debugf(string(body))

	var uploadPhotoStatus *uploadPhotoStatusStruct
	err = json.Unmarshal(body, &uploadPhotoStatus)
	if err != nil {
		return "", err
	}

	if uploadPhotoStatus.MediaID == "" {
		return "", errors.New(uploadPhotoStatus.Message)
	}

	return uploadPhotoStatus.MediaID, nil
}
