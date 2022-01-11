package main

import (
	"encoding/json"
	"fmt"
	"github.com/imroc/req"
	"github.com/robfig/cron/v3"
	"log"
	"net"
	"regexp"
	"searchproxy/fram/config"
	"searchproxy/fram/utils"
	"strconv"
	"strings"
)

type pushmsg struct {
	Pushurl   string `json:"pushurl"`
	Getqueues string `json:"getqueues"`
}

func push(topic, msg, url string) {
	_, err := req.Post(url, req.BodyJSON(req.Param{
		"delivery_mode":    "1",
		"headers":          req.Param{},
		"name":             "amq.default",
		"payload":          msg,
		"payload_encoding": "string",
		"properties":       req.Param{"delivery_mode": 1, "headers": req.Param{}},
		"props":            req.Param{},
		"routing_key":      topic,
		"vhost":            "/",
	}), req.Header{"x-vhost": "",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36",
	})
	if err != nil {
		panic(err)
	}
}

func getcount(topic, url string) int64 {
	resp, err := req.Get(url + topic)
	if err != nil {
		return 0
	}
	var data map[string]interface{}
	err = resp.ToJSON(&data)
	if err != nil {
		return 0
	}
	if messages, ok := data["messages"]; ok {
		i64, err := strconv.ParseInt(fmt.Sprintf("%v", messages), 10, 64)
		if err != nil {
			return 0
		}
		return i64
	}
	return 0
}

func ipfilter(ip string) bool {
	matchString, err := regexp.MatchString("((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})(\\.((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})){3}", ip)
	if err != nil {
		return false
	}
	if matchString {

		addr, err := net.ResolveIPAddr("ip", "google.com")
		if err != nil {
			return false
		}
		if addr.String() == ip {
			return false
		}

		addr, err = net.ResolveIPAddr("ip", "baidu.com")
		if err != nil {
			return false
		}
		if addr.String() == ip {
			return false
		}

		addr, err = net.ResolveIPAddr("ip", "www.google.com")
		if err != nil {
			return false
		}
		if addr.String() == ip {
			return false
		}

		addr, err = net.ResolveIPAddr("ip", "www.baidu.com")
		if err != nil {
			return false
		}
		if addr.String() == ip {
			return false
		}

		ipreg := `^10\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])$`
		matchString, err = regexp.MatchString(ipreg, ip)
		if err != nil {
			return false
		}
		if matchString {
			return false
		}

		ipreg = `^172\.(1[6789]|2[0-9]|3[01])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])$`
		matchString, err = regexp.MatchString(ipreg, ip)
		if err != nil {
			return false
		}
		if matchString {
			return false
		}

		ipreg = `^192\.168\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])$`
		matchString, err = regexp.MatchString(ipreg, ip)
		if err != nil {
			return false
		}
		if matchString {
			return false
		}
		if ip == "1.1.1.1" || ip == "8.8.8.8" || ip == "114.114.114.114" || ip == "222.222.222.222" || strings.Split(ip, ".")[0] == "127" || strings.Split(ip, ".")[0] == "0" {
			return false
		}
		return true
	}
	return false
}

func taskpush(m *pushmsg) {
	for i := utils.Ip2Int64("1.0.0.0"); i < utils.Ip2Int64("255.255.255.255"); i++ {
		if !ipfilter(utils.Int64ToIp(i)) {
			log.Println(utils.Int64ToIp(i), "continue")
			continue
		}
		log.Println(utils.Int64ToIp(i))
		marshal, err := json.Marshal(map[string]interface{}{
			"ip":   utils.Int64ToIp(i),
			"rate": 1000,
		})
		if err != nil {
			continue
		}
		go push("scanproxy", string(marshal), m.Pushurl)
	}
}

func main() {
	var msg pushmsg
	config.Install().Get("mq", &msg)
	taskpush(&msg)
	c := cron.New() // 新建一个定时任务对象
	c.AddFunc("*/10 * * * *", func() {
		if getcount("scanproxy", msg.Getqueues) == 0 {
			taskpush(&msg)
		}
	}) // 给对象增加定时任务
	c.Start()
	select {}
}
