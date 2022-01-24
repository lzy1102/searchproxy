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
	"net/url"
	"strconv"
	"strings"
	"time"
)

// proxy,google,protocol
func scanproxy(ip string, port int) (bool, bool, string) {
	//req.Client().Jar, _ = cookiejar.New(nil)
	//trans, _ := req.Client().Transport.(*http.Transport)
	//trans.TLSHandshakeTimeout = time.Duration(myinfo.timeout) * time.Second
	//trans.DisableKeepAlives = true
	//trans.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("%v:%v", ip, port), nil, proxy.Direct)
	if err != nil {
		return false, false, ""
	}
	client := &http.Client{
		Timeout: time.Duration(myinfo.timeout) * time.Second,
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
		Timeout: time.Duration(myinfo.timeout) * time.Second,
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
		Timeout: time.Duration(myinfo.timeout) * time.Second,
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
		Timeout: time.Duration(myinfo.timeout) * time.Second,
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
		Timeout: time.Duration(myinfo.timeout) * time.Second,
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

func proxysocks5google(ip string, port int, c chan interface{}) {
	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("%v:%v", ip, port), nil, proxy.Direct)
	if err != nil {
		c <- map[string]interface{}{
			"ip":       ip,
			"port":     port,
			"proxy":    false,
			"google":   false,
			"protocol": "",
		}
	}
	client := &http.Client{
		Timeout: time.Duration(myinfo.timeout) * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
				return dialer.Dial(network, addr)
			},
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		},
	}
	r, err := req.Get("https://www.google.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("google", r.Response().Body) {
		c <- map[string]interface{}{
			"ip":       ip,
			"port":     port,
			"proxy":    true,
			"google":   true,
			"protocol": "socks5",
		}
	} else {
		c <- map[string]interface{}{
			"ip":       ip,
			"port":     port,
			"proxy":    false,
			"google":   false,
			"protocol": "",
		}
	}
}
func proxysocks5baidu(ip string, port int, c chan interface{}) {
	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("%v:%v", ip, port), nil, proxy.Direct)
	if err != nil {
		c <- map[string]interface{}{
			"ip":       ip,
			"port":     port,
			"proxy":    false,
			"google":   false,
			"protocol": "",
		}
	}
	client := &http.Client{
		Timeout: time.Duration(myinfo.timeout) * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
				return dialer.Dial(network, addr)
			},
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		},
	}
	r, err := req.Get("https://www.baidu.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("百度一下，你就知道", r.Response().Body) {
		c <- map[string]interface{}{
			"ip":       ip,
			"port":     port,
			"proxy":    true,
			"google":   false,
			"protocol": "socks5",
		}
	} else {
		c <- map[string]interface{}{
			"ip":       ip,
			"port":     port,
			"proxy":    false,
			"google":   false,
			"protocol": "",
		}
	}
}

func proxyhttpgoogle(ip string, port int, c chan interface{}) {
	urlproxy, _ := url.Parse(fmt.Sprintf("http://%v:%v", ip, port))
	client := &http.Client{
		Timeout: time.Duration(myinfo.timeout) * time.Second,
		Transport: &http.Transport{
			Proxy:             http.ProxyURL(urlproxy),
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		}}
	r, err := req.Get("https://www.google.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("google", r.Response().Body) {
		c <- map[string]interface{}{
			"ip":       ip,
			"port":     port,
			"proxy":    true,
			"google":   true,
			"protocol": "http",
		}
	} else {
		c <- map[string]interface{}{
			"ip":       ip,
			"port":     port,
			"proxy":    false,
			"google":   false,
			"protocol": "",
		}
	}
}
func proxyhttpbaidu(ip string, port int, c chan interface{}) {
	urlproxy, _ := url.Parse(fmt.Sprintf("http://%v:%v", ip, port))
	client := &http.Client{
		Timeout: time.Duration(myinfo.timeout) * time.Second,
		Transport: &http.Transport{
			Proxy:             http.ProxyURL(urlproxy),
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		}}
	r, err := req.Get("https://www.baidu.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("百度一下，你就知道", r.Response().Body) {
		c <- map[string]interface{}{
			"ip":       ip,
			"port":     port,
			"proxy":    true,
			"google":   false,
			"protocol": "http",
		}
	} else {
		c <- map[string]interface{}{
			"ip":       ip,
			"port":     port,
			"proxy":    false,
			"google":   false,
			"protocol": "",
		}
	}
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
	datachan := make(chan interface{}, 4)
	//proxystatus, isgoogle, protocol := false, false, ""
	//proxystatus, isgoogle, protocol = scanproxy(ip, atoi)
	go proxysocks5google(ip, atoi, datachan)
	go proxysocks5baidu(ip, atoi, datachan)
	go proxyhttpgoogle(ip, atoi, datachan)
	go proxyhttpbaidu(ip, atoi, datachan)
	for i := 0; i < 4; i++ {
		tmp := <-datachan
		if proxystatus, ok := tmp.(map[string]interface{})["proxy"]; ok && proxystatus.(bool) {
			result = append(result, tmp)
		}
	}
	//if proxystatus {
	//	result = append(result, map[string]interface{}{
	//		"ip":       ip,
	//		"port":     port,
	//		"proxy":    proxystatus,
	//		"google":   isgoogle,
	//		"protocol": protocol,
	//	})
	//}

	return result
}

type info struct {
	ip      string
	port    string
	timeout int
	out     string
}

var myinfo info

func init() {
	flag.StringVar(&myinfo.ip, "ip", "127.0.0.1", "target ip")
	flag.StringVar(&myinfo.port, "port", "1080", "port list")
	flag.IntVar(&myinfo.timeout, "timeout", 5, "time out sec")
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
