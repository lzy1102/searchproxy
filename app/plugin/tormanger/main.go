package main

import (
	"flag"
	"fmt"
	"github.com/go-ping/ping"
	"github.com/imroc/req"
	"io/ioutil"
	"time"
)

func settorrc(protocol, ip, port string) {
	if protocol == "socks5" {
		ioutil.WriteFile("/etc/tor/torrc", []byte(fmt.Sprintf("Socks5Proxy %v:%v", ip, port)), 777)
	}

}

func pinggoogle() bool {
	pinger, err := ping.NewPinger("google.com")
	if err != nil {
		return false
	}
	pinger.Count = 3
	pinger.Timeout = 20000 * time.Millisecond
	pinger.SetPrivileged(true)
	_ = pinger.Run() // blocks until finished
	stats := pinger.Statistics()
	if stats.PacketsRecv >= 1 {
		return true
	} else {
		return false
	}
}

func main() {
	addr := flag.String("addr", "http://proxy.xinjing123.top:8080/api/get/list", "")
	flag.Parse()
	for {
		if !pinggoogle() {
			req.Get(*addr, req.QueryParam{
				"google":   true,
				"protocol": "socks5",
				"skip":     0,
				"limit":    1,
			})
		}
	}
}
