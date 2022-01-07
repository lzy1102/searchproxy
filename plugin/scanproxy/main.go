package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/cheggaaa/pb/v3"
	"github.com/imroc/req"
	"github.com/tevino/tcp-shaker"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// proxy,google,protocol
func scanproxy(ip, port string) (bool, bool, string) {
	req.Client().Jar, _ = cookiejar.New(nil)
	trans, _ := req.Client().Transport.(*http.Transport)
	trans.TLSHandshakeTimeout = 1 * time.Second
	trans.DisableKeepAlives = true
	trans.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("%v:%v", ip, port), nil, proxy.Direct)
	if err != nil {
		return false, false, ""
	}
	client := &http.Client{
		Timeout: 1 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
				return dialer.Dial(network, addr)
			}}}
	r, err := req.Get("http://www.google.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("google", r.Response().Body) {
		return true, true, "socks5"
	}
	r, err = req.Get("http://www.baidu.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("百度一下，你就知道", r.Response().Body) {
		return true, false, "socks5"
	}

	urlproxy, _ := url.Parse(fmt.Sprintf("http://%v:%v", ip, port))
	client = &http.Client{
		Timeout: 1 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(urlproxy),
		}}
	r, err = req.Get("http://www.google.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("google", r.Response().Body) {
		return true, true, "http"
	}
	urlproxy, _ = url.Parse(fmt.Sprintf("https://%v:%v", ip, port))
	client = &http.Client{
		Timeout: 1 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(urlproxy),
		}}
	r, err = req.Get("http://www.google.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("google", r.Response().Body) {
		return true, true, "https"
	}
	urlproxy, _ = url.Parse(fmt.Sprintf("http://%v:%v", ip, port))
	client = &http.Client{
		Timeout: 1 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(urlproxy),
		}}
	r, err = req.Get("http://www.baidu.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil && r.Response().StatusCode == 200 && checktitle("百度一下，你就知道", r.Response().Body) {
		return true, false, "http"
	}
	urlproxy, _ = url.Parse(fmt.Sprintf("https://%v:%v", ip, port))
	client = &http.Client{
		Timeout: 1 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(urlproxy),
		}}
	r, err = req.Get("http://www.baidu.com", client, req.Header{
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

func socketdial(ip, port string) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", ip, port), 1*time.Second)
	if conn!=nil {
		conn.Close()
	}
	if err != nil {
		return false
	}
	return true
}

func tcpshaker(ip,port string) bool {
	c := tcp.NewChecker()
	ctx, stopChecker := context.WithCancel(context.Background())
	defer stopChecker()
	go func() {
		if err := c.CheckingLoop(ctx); err != nil {
			fmt.Println("checking loop stopped due to fatal error: ", err)
		}
	}()

	<-c.WaitReady()

	timeout := time.Second * 1
	err := c.CheckAddr(fmt.Sprintf("%v:%v",ip,port), timeout)
	if err==nil {
		return true
	}
	return false
}

func scan(ip string, rate int) (result []interface{}) {
	ratechan := make(chan interface{}, rate) // 控制任务并发的chan
	datachan := make(chan interface{}, 0)
	bar := pb.StartNew(65535)
	for i := 1; i < 65535; i++ {
		ratechan <- struct{}{} // 作用类似于waitgroup.Add(1)
		bar.Increment()
		go func(host, port string) {
			//portstatus := socketdial(host, port)
			portstatus := tcpshaker(host, port)
			var proxystatus, isgoogle bool
			var protocol string
			if portstatus {
				log.Println(port,portstatus)
				proxystatus, isgoogle, protocol = scanproxy(host, port)
			}
			<-ratechan // 执行完毕，释放资源
			datachan <- map[string]interface{}{
				"ip":       host,
				"port":     port,
				"status":   portstatus,
				"proxy":    proxystatus,
				"google":   isgoogle,
				"protocol": protocol,
			}
		}(ip, fmt.Sprintf("%v", i))
	}
	for i := 1; i < 65535; i++ {
		tmp := <-datachan
		if proxystatus, ok := tmp.(map[string]interface{})["proxy"]; ok && proxystatus.(bool) {
			result = append(result, tmp)
		}
	}
	bar.Finish()
	return result
}

type info struct {
	ip   string
	rate int
	out  string
}

var myinfo info

func init() {
	flag.StringVar(&myinfo.ip, "ip", "127.0.0.1", "target ip")
	flag.IntVar(&myinfo.rate, "rate", 100, "thread number")
	flag.StringVar(&myinfo.out, "out", "out.json", "out json file name")
	flag.Parse()
	if myinfo.ip == "" {
		log.Fatalln("ip is err")
	}
}

func main() {
	result := scan(myinfo.ip, myinfo.rate)
	out, err := json.Marshal(result)
	if err == nil {
		_ = ioutil.WriteFile(myinfo.out, out, 0777)
	}
}
