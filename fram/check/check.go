package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/imroc/req"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"searchproxy/fram/info"
	"searchproxy/fram/utils"
	"time"
)

func CheckProxy(ip, port string) {
	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("%v:%v", ip, port), nil, proxy.Direct)
	utils.FatalAssert(err)
	httpTransport := &http.Transport{DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
		return dialer.Dial(network, addr)
	}}
	client := &http.Client{Timeout: 5 * time.Second, Transport: httpTransport}
	//utils.FatalAssert(req.SetProxyUrl(fmt.Sprintf("socks5://%v:%v", ip, port)))
	r, err := req.Get("http://google.com", client, req.Header{
		`User-Agent`: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
	})
	if err == nil {
		if r.Response().StatusCode == 200 {
			log.Println(ip, port)
		}
	}
}

func test() {
	out, err := ioutil.ReadFile("fofa.json")
	utils.FatalAssert(err)
	var data []info.Data
	utils.FatalAssert(json.Unmarshal(out, &data))
	for _, i2 := range data {
		log.Println(i2.Ip, utils.Ip2Int64(i2.Ip))
		CheckProxy(i2.Ip, i2.Port)
	}
}

func main() {

}
