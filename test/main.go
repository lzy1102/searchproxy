package main

import (
	"flag"
	"github.com/imroc/req"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

func createtopic(topic, url string) {

	headers := req.Header{
		"Connection": "keep-alive",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36",
	}

	data := req.Param{
		"vhost":       "/",
		"name":        topic,
		"durable":     "true",
		"auto_delete": false,
		"arguments":   req.Param{"x-queue-type": "classic"},
	}

	response, _ := req.Put("http://proxy:3f0c1304c3865ea6@172.16.30.190:15672/api/queues/%2F/"+topic, headers, req.BodyJSON(data))
	log.Println(response.Response().Status)
}

func ipfilter(ip string) bool {
	matchString, err := regexp.MatchString("((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})(\\.((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})){3}", ip)
	if err != nil {
		log.Println("非法IP")
		return false
	}
	if matchString {

		addr, err := net.ResolveIPAddr("ip", "google.com")
		if err != nil || addr.String() == ip {
			log.Println("是 google")
			return false
		}

		addr, err = net.ResolveIPAddr("ip", "baidu.com")
		if err != nil || addr.String() == ip {
			log.Println("是 百度")
			return false
		}

		addr, err = net.ResolveIPAddr("ip", "www.google.com")
		if err != nil || addr.String() == ip {
			log.Println("是 google")
			return false
		}

		addr, err = net.ResolveIPAddr("ip", "www.baidu.com")
		if err != nil || addr.String() == ip {
			log.Println("是 百度")
			return false
		}

		ipreg := `^10\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])$`
		matchString, err = regexp.MatchString(ipreg, ip)
		if err != nil || matchString {
			log.Println("是 10段")
			return false
		}

		ipreg = `^172\.(1[6789]|2[0-9]|3[01])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])$`
		matchString, err = regexp.MatchString(ipreg, ip)
		if err != nil || matchString {
			log.Println("是 172段")
			return false
		}

		ipreg = `^192\.168\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])$`
		matchString, err = regexp.MatchString(ipreg, ip)
		if err != nil || matchString {
			log.Println("是 192段")
			return false
		}
		if strings.Split(ip, ".")[0] == "127" || strings.Split(ip, ".")[0] == "0" {
			log.Println("是DNS")
			return false
		}
		return true
	}
	return false
}

func main() {
	var cfgaddr string
	flag.StringVar(&cfgaddr, "cfgaddr", "", "")
	flag.Parse()
	for i, arg := range os.Args {
		log.Println(i, arg)
	}
	log.Println(cfgaddr)
}
