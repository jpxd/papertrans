package main

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/crypto/ssh"
)

const userAgent = "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"

// WebClient ...
type WebClient struct {
	client    *http.Client
	cookieJar *cookiejar.Jar
}

func CreateWebClient() *WebClient {
	wc := WebClient{}
	wc.cookieJar, _ = cookiejar.New(nil)
	wc.client = &http.Client{
		Jar: wc.cookieJar,
	}
	return &wc
}

func CreateTunneledWebClient(sshConn *ssh.Client) *WebClient {
	wc := WebClient{}
	wc.cookieJar, _ = cookiejar.New(nil)
	wc.client = &http.Client{
		Jar: wc.cookieJar,
		Transport: &http.Transport{
			DialContext: createSSHDialContext(sshConn),
		},
	}
	return &wc
}

func (wc *WebClient) PerformRequest(req *http.Request) *http.Response {
	req.Header.Set("User-Agent", userAgent)

	resp, err := wc.client.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
}

func (wc *WebClient) Get(url string) *http.Response {
	req, _ := http.NewRequest("GET", url, nil)
	return wc.PerformRequest(req)
}

func (wc *WebClient) PostForm(url string, data url.Values) *http.Response {
	req, _ := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return wc.PerformRequest(req)
}
