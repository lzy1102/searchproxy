package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// proxy,google,protocol
func scanproxy(ip string, port int) (bool, bool, string) {
	req.Client().Jar, _ = cookiejar.New(nil)
	trans, _ := req.Client().Transport.(*http.Transport)
	trans.TLSHandshakeTimeout = 5 * time.Second
	trans.DisableKeepAlives = true
	trans.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("%v:%v", ip, port), nil, proxy.Direct)
	if err != nil {
		return false, false, ""
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
				return dialer.Dial(network, addr)
			}}}
	r, err := req.Get("https://www.google.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("google", r.Response().Body) {
		return true, true, "socks5"
	}
	r, err = req.Get("https://www.baidu.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("百度一下，你就知道", r.Response().Body) {
		return true, false, "socks5"
	}

	urlproxy, _ := url.Parse(fmt.Sprintf("http://%v:%v", ip, port))
	client = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(urlproxy),
		}}
	r, err = req.Get("https://www.google.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("google", r.Response().Body) {
		return true, true, "http"
	}
	urlproxy, _ = url.Parse(fmt.Sprintf("https://%v:%v", ip, port))
	client = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(urlproxy),
		}}
	r, err = req.Get("https://www.google.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("google", r.Response().Body) {
		return true, true, "https"
	}
	urlproxy, _ = url.Parse(fmt.Sprintf("http://%v:%v", ip, port))
	client = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(urlproxy),
		}}
	r, err = req.Get("https://www.baidu.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("百度一下，你就知道", r.Response().Body) {
		return true, false, "http"
	}
	urlproxy, _ = url.Parse(fmt.Sprintf("https://%v:%v", ip, port))
	client = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(urlproxy),
		}}
	r, err = req.Get("https://www.baidu.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("百度一下，你就知道", r.Response().Body) {
		return true, false, "https"
	}

	return false, false, ""
}

func checktitle(title string, body io.Reader) bool {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return false
	}
	findtitle := doc.Find("title").Text()
	return strings.ToLower(title) == strings.ToLower(findtitle)
}

func synscan(ip, port string) (result []interface{}) {
	atoi, err := strconv.Atoi(port)
	if err != nil {
		return nil
	}
	proxystatus, isgoogle, protocol := false, false, ""
	proxystatus, isgoogle, protocol = scanproxy(ip, atoi)
	if proxystatus {
		result = append(result, map[string]interface{}{
			"ip":       ip,
			"port":     port,
			"proxy":    proxystatus,
			"google":   isgoogle,
			"protocol": protocol,
		})
	}

	return result
}

type info struct {
	ip   string
	port string
	out  string
}

var myinfo info

func init() {
	flag.StringVar(&myinfo.ip, "ip", "127.0.0.1", "target ip")
	flag.StringVar(&myinfo.port, "port", "1080", "port list")
	flag.StringVar(&myinfo.out, "out", "out.json", "out json file name")
	flag.Parse()
	if myinfo.ip == "" {
		log.Fatalln("ip is err")
	}
}

func main() {
	var result []interface{}
	result = synscan(myinfo.ip, myinfo.port)
	if result != nil && len(result) > 0 {
		out, err := json.Marshal(result)
		if err == nil {
			_ = ioutil.WriteFile(myinfo.out, out, 0777)
		}
	}
}
