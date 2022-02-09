package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/imroc/req"
	"golang.org/x/net/proxy"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

func main() {
	addr := flag.String("addr", "http://81.68.237.86:8080/api/get/list", "")
	deleteaddr := flag.String("del", "http://81.68.237.86:8080/api/post/delete", "")
	flag.Parse()
	for {
		skip := 0
		var offlist []interface{}
		for true {
			resp, err := req.Get(*addr, req.QueryParam{
				"google":   true,
				"protocol": "socks5",
				"limit":    10,
				"skip":     skip,
			})
			if err != nil {
				return
			}
			var result []interface{}
			resp.ToJSON(&result)
			if len(result) < 10 {
				break
			}
			for _, i2 := range result {
				ip := i2.(map[string]interface{})["ip"]
				port := i2.(map[string]interface{})["port"]
				oid := i2.(map[string]interface{})["_id"]
				dialer, _ := proxy.SOCKS5("tcp", fmt.Sprintf("%v:%v", ip, port), nil, proxy.Direct)

				client := &http.Client{
					Timeout: time.Second * 10,
					Transport: &http.Transport{
						DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
							return dialer.Dial(network, addr)
						},
						DisableKeepAlives: true,
						TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
					},
				}
				_, err := req.Get("https://www.google.com", client, req.Header{
					`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
				})
				if err != nil {
					log.Println(ip, port, "off")
					offlist = append(offlist, oid)
				} else {
					log.Println(ip, port, "open")
				}
			}
			skip += 10
		}
		skip = 0
		for true {
			resp, err := req.Get(*addr, req.QueryParam{
				"google":   false,
				"protocol": "socks5",
				"limit":    10,
				"skip":     skip,
			})
			if err != nil {
				return
			}
			var result []interface{}
			resp.ToJSON(&result)
			if len(result) < 10 {
				break
			}
			for _, i2 := range result {
				ip := i2.(map[string]interface{})["ip"]
				port := i2.(map[string]interface{})["port"]
				oid := i2.(map[string]interface{})["_id"]
				dialer, _ := proxy.SOCKS5("tcp", fmt.Sprintf("%v:%v", ip, port), nil, proxy.Direct)

				client := &http.Client{
					Timeout: time.Second * 10,
					Transport: &http.Transport{
						DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
							return dialer.Dial(network, addr)
						},
						DisableKeepAlives: true,
						TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
					},
				}
				_, err := req.Get("https://www.baidu.com", client, req.Header{
					`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
				})
				if err != nil {
					log.Println(ip, port, "off")
					offlist = append(offlist, oid)
				} else {
					log.Println(ip, port, "open")
				}
			}
			skip += 10
		}

		skip = 0
		for true {
			resp, err := req.Get(*addr, req.QueryParam{
				"google":   false,
				"protocol": "http",
				"limit":    10,
				"skip":     skip,
			})
			if err != nil {
				return
			}
			var result []interface{}
			resp.ToJSON(&result)
			if len(result) < 10 {
				break
			}
			for _, i2 := range result {
				ip := i2.(map[string]interface{})["ip"]
				port := i2.(map[string]interface{})["port"]
				oid := i2.(map[string]interface{})["_id"]

				urlproxy, _ := url.Parse(fmt.Sprintf("http://%v:%v", ip, port))
				client := &http.Client{
					Timeout: time.Second * 10,
					Transport: &http.Transport{
						Proxy:             http.ProxyURL(urlproxy),
						DisableKeepAlives: true,
						TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
					}}
				_, err := req.Get("https://www.baidu.com", client, req.Header{
					`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
				})
				if err != nil {
					log.Println(ip, port, "off")
					offlist = append(offlist, oid)
				} else {
					log.Println(ip, port, "open")
				}
			}
			skip += 10
		}

		skip = 0
		for true {
			resp, err := req.Get(*addr, req.QueryParam{
				"google":   true,
				"protocol": "http",
				"limit":    10,
				"skip":     skip,
			})
			if err != nil {
				return
			}
			var result []interface{}
			resp.ToJSON(&result)
			if len(result) < 10 {
				break
			}
			for _, i2 := range result {
				ip := i2.(map[string]interface{})["ip"]
				port := i2.(map[string]interface{})["port"]
				oid := i2.(map[string]interface{})["_id"]

				urlproxy, _ := url.Parse(fmt.Sprintf("http://%v:%v", ip, port))
				client := &http.Client{
					Timeout: time.Second * 10,
					Transport: &http.Transport{
						Proxy:             http.ProxyURL(urlproxy),
						DisableKeepAlives: true,
						TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
					}}
				_, err := req.Get("https://www.google.com", client, req.Header{
					`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
				})
				if err != nil {
					log.Println(ip, port, "off")
					offlist = append(offlist, oid)
				} else {
					log.Println(ip, port, "open")
				}
			}
			skip += 10
		}

		for _, i2 := range offlist {
			req.Post(*deleteaddr, req.Param{
				"id": i2,
			})
		}
	}
}
