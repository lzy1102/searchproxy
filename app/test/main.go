package main

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
)

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
	log.Println(ipfilter(strings.Split("172.16.10.1/24", "/")[0]))
	for a := 1; a < 255; a++ {
		for b := 0; b < 255; b++ {
			for c := 0; c < 255; c++ {
				ip := fmt.Sprintf("%v.%v.%v.1/24", a, b, c)
				if !ipfilter(strings.Split(ip, "/")[0]) {
					log.Println(ip)
				} else {
					fmt.Println(ip)
				}
			}
		}
	}
}
