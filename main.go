package main

import (
	"log"
	"net/http"
	"bytes"
	"io"
	"io/ioutil"
	"github.com/qiniu/api.v7/auth/qbox"
	"encoding/json"
	"fmt"
	"crypto/hmac"
	"crypto/sha1"
	"github.com/qiniu/x/bytes.v7/seekable"
	"encoding/base64"

	"os"
)

var (
	accessKey1 = ""
	secretKey1 = ""
	urlsToRefresh []string
	dirsToRefresh []string
)

// RefreshReq 为缓存刷新请求内容
type RefreshReq struct {
	Urls []string `json:"urls"`
	Dirs []string `json:"dirs"`
}

// RefreshResp 缓存刷新响应内容
type RefreshResp struct {
	Code          int      `json:"code"`
	Error         string   `json:"error"`
	RequestID     string   `json:"requestId,omitempty"`
	InvalidUrls   []string `json:"invalidUrls,omitempty"`
	InvalidDirs   []string `json:"invalidDirs,omitempty"`
	URLQuotaDay   int      `json:"urlQuotaDay,omitempty"`
	URLSurplusDay int      `json:"urlSurplusDay,omitempty"`
	DirQuotaDay   int      `json:"dirQuotaDay,omitempty"`
	DirSurplusDay int      `json:"dirSurplusDay,omitempty"`
}

func main()  {


	for i := 0; i < len(os.Args); i++ {
		switch os.Args[i] {

		case "-l":
			for os.Args[i+1][0] != '-' {
				urlsToRefresh = append(urlsToRefresh, os.Args[i+1])
				if i < len(os.Args)-2 {
					i++
				} else {
					break
				}
			}
		case "-d":
			for os.Args[i+1][0] != '-' {
				dirsToRefresh = append(dirsToRefresh, os.Args[i+1])
				if i < len(os.Args)-2 {
					i++
				} else {
					break
				}
			}
		case "-h":
			log.Println("\n-l	刷新链接 \n-d	刷新目录")
			return

		}
	}


	tune_api := "http://fusion.qiniuapi.com/v2/tune/refresh"

	if len(urlsToRefresh) > 100 {
		log.Println("urls count exceeds the limit of 100")
		return
	}

	log.Println(urlsToRefresh)
	log.Println(dirsToRefresh)

	reqBody := RefreshReq{
		Urls: urlsToRefresh,
		Dirs: dirsToRefresh,
	}

	mac := qbox.NewMac(accessKey1, secretKey1)

	reqData, _ := json.Marshal(reqBody)


	req, reqErr := http.NewRequest("POST", tune_api, bytes.NewReader(reqData))
	if reqErr != nil {
		log.Println(reqErr)
		return
	}

	h := hmac.New(sha1.New, mac.SecretKey)
	u := req.URL
	data := u.Path
	if u.RawQuery != "" {
		data += "?" + u.RawQuery
	}
	io.WriteString(h, data+"\n")

	if incBody(req) {
		s2, err2 := seekable.New(req)
		if err2 != nil {
			log.Println(err2)
		}
		h.Write(s2.Bytes())
	}

	sign := base64.URLEncoding.EncodeToString(h.Sum(nil))
	accessToken := fmt.Sprintf("%s:%s", mac.AccessKey, sign)


	req.Header.Add("Authorization", "QBox "+accessToken)
	req.Header.Add("Content-Type", "application/json")


	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(reqErr)
		return
	}
	defer resp.Body.Close()

	var rfr  RefreshResp

	umErr := json.Unmarshal(body, &rfr)
	if umErr != nil {
		log.Println(reqErr)

	}
	log.Println(rfr)

}

func incBody(req *http.Request) bool {
	return req.Body != nil &&
		req.Header.Get("Content-Type") == "application/x-www-form-urlencoded"
}