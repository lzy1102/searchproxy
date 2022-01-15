//go:build linux
// +build linux

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/cheggaaa/pb/v3"
	"github.com/dean2021/go-masscan"
	"github.com/imroc/req"
	"github.com/tevino/tcp-shaker"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"searchproxy/app/fram/utils"
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

func socketdial(ip, port string) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", ip, port), 5*time.Second)
	if conn != nil {
		conn.Close()
	}
	if err != nil {
		return false
	}
	return true
}

func tcpshaker(ip string, port int) bool {
	c := tcp.NewChecker()
	ctx, stopChecker := context.WithCancel(context.TODO())
	defer stopChecker()
	go func() {
		if err := c.CheckingLoop(ctx); err != nil {
			fmt.Println("checking loop stopped due to fatal error: ", err)
		}
	}()

	<-c.WaitReady()

	timeout := 5 * time.Second
	err := c.CheckAddr(fmt.Sprintf("%v:%v", ip, port), timeout)
	if err == nil {
		return true
	}
	return false
}

func synscan(ip, ports string, rate int) (result []interface{}) {
	ratechan := make(chan interface{}, rate) // 控制任务并发的chan
	datachan := make(chan interface{}, 0)
	iplist := getallip(ip)
	var portlist []int
	for _, v := range strings.Split(strings.TrimSpace(ports), ",") {
		if strings.Contains(strings.TrimSpace(v), "-") {
			tmp := strings.Split(v, "-")
			startport, err := strconv.Atoi(tmp[0])
			endport, err := strconv.Atoi(tmp[len(tmp)-1])
			if err != nil {
				continue
			}
			for i := startport; i <= endport; i++ {
				portlist = append(portlist, i)
			}
		} else {
			atoi, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			portlist = append(portlist, atoi)
		}
	}
	bar := pb.StartNew(len(portlist) * len(iplist))
	for _, s := range iplist {
		for _, p := range portlist {
			ratechan <- struct{}{} // 作用类似于waitgroup.Add(1)
			bar.Increment()
			go func(host string, port int) {
				portstatus := tcpshaker(host, port)
				proxystatus, isgoogle, protocol := false, false, ""
				if portstatus == true {
					log.Println("port", port, "status", portstatus)
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
			}(s, p)
		}
	}
	for _ = range iplist {
		for range portlist {
			tmp := <-datachan
			if proxystatus, ok := tmp.(map[string]interface{})["proxy"]; ok && proxystatus.(bool) {
				result = append(result, tmp)
			}
		}
	}
	bar.Finish()
	return result
}

func masscaner(target, ports, rate string) []interface{} {
	m := masscan.New()
	// masscan可执行文件路径,默认不需要设置
	//m.SetSystemPath("D:\\Program Files\\masscan/masscan.exe")
	// 扫描端口范围
	m.SetPorts(ports)
	// 扫描IP范围
	m.SetRanges(target)
	// 扫描速率
	m.SetRate(rate)
	// 隔离扫描名单
	//m.SetExclude("127.0.0.1")
	// 开始扫描
	err := m.Run()
	if err != nil {
		fmt.Println("scanner failed:", err)
		return nil
	}
	// 解析扫描结果
	results, err := m.Parse()
	if err != nil {
		fmt.Println("Parse scanner result:", err)
		return nil
	}
	data := make([]interface{}, 0)
	count := 0
	datachan := make(chan interface{}, 0)
	for _, result := range results {
		for _, port := range result.Ports {
			if port.State.State == "open" {
				count++
				atoi, _ := strconv.Atoi(port.Portid)
				go func(host string, port int) {
					proxystatus, isgoogle, protocol := false, false, ""
					proxystatus, isgoogle, protocol = scanproxy(host, port)
					datachan <- map[string]interface{}{
						"ip":       host,
						"port":     port,
						"status":   true,
						"proxy":    proxystatus,
						"google":   isgoogle,
						"protocol": protocol,
					}
				}(result.Address.Addr, atoi)
			}
		}
	}
	for i := 0; i < count; i++ {
		tmp := <-datachan
		if proxystatus, ok := tmp.(map[string]interface{})["proxy"]; ok && proxystatus.(bool) {
			data = append(data, tmp)
		}
	}
	return data
}

type info struct {
	ip     string
	rate   int
	scaner string
	ports  string
	out    string
}

var myinfo info

func init() {
	flag.StringVar(&myinfo.ip, "ip", "127.0.0.1", "target ip")
	flag.IntVar(&myinfo.rate, "rate", 1000, "thread number")
	flag.StringVar(&myinfo.scaner, "scaner", "syn", "scan name")
	flag.StringVar(&myinfo.ports, "ports", "1080", "port list")
	flag.StringVar(&myinfo.out, "out", "out.json", "out json file name")
	flag.Parse()
	if myinfo.ip == "" {
		log.Fatalln("ip is err")
	}
}

func getallip(ip string) []string {
	iplist := make([]string, 0)
	if strings.Contains(ip, "/") {
		tmp := strings.Split(ip, "/")
		atoi, err := strconv.Atoi(tmp[1])
		if err != nil {
			return []string{}
		}
		maxhost := int(math.Pow(float64(2), float64(32-atoi+1))) - 2
		minip := tmp[0]
		tmpip := strings.Split(tmp[0], ".")
		hostid, err := strconv.Atoi(tmpip[3])
		if err != nil {
			return []string{}
		}
		maxip := fmt.Sprintf("%v.%v.%v.%v", tmpip[0], tmpip[1], tmpip[2], hostid+maxhost)
		iplist = utils.GetIpAll(minip, maxip)
	} else {
		iplist = append(iplist, ip)
	}
	return iplist
}

func main() {
	var result []interface{}

	if myinfo.scaner == "syn" {
		result = synscan(myinfo.ip, myinfo.ports, myinfo.rate)
	} else if myinfo.scaner == "masscan" {
		result = masscaner(myinfo.ip, myinfo.ports, fmt.Sprintf("%v", myinfo.rate))
	}

	if result != nil && len(result) > 0 {
		out, err := json.Marshal(result)
		if err == nil {
			_ = ioutil.WriteFile(myinfo.out, out, 0777)
		}
	}
}
