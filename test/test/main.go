package main

import (
	"crypto/tls"
	"github.com/imroc/req"
	"log"
	"net/http"
	"net/url"
	"time"
)

func main() {
	urlproxy, _ := url.Parse("http://206.253.164.101:80")
	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Proxy:             http.ProxyURL(urlproxy),
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		}}
	r, err := req.Get("http://www.baidu.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err != nil {
		panic(err)
	}
	log.Println(r.String())
}
