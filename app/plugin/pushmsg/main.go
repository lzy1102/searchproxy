package main

import (
	"encoding/json"
	"fmt"
	"github.com/imroc/req"
	"github.com/robfig/cron/v3"
	"log"
	"math/rand"
	"net"
	"regexp"
	"searchproxy/app/fram/config"
	"searchproxy/app/fram/db"
	"searchproxy/app/fram/logs"
	"strconv"
	"strings"
	"time"
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

func istopic(topic, url string) bool {
	resp, err := req.Get(url)
	if err != nil {
		return false
	}
	var data []interface{}
	err = resp.ToJSON(&data)
	if err != nil {
		return false
	}
	for _, datum := range data {
		tmp := datum.(map[string]interface{})
		if name, ok := tmp["name"]; ok && name.(string) == topic {
			return true
		}
	}
	return false
}

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
		//"arguments":   req.Param{"x-queue-type": "classic"},
	}

	req.Put(url+topic, headers, req.BodyJSON(data))
	//log.Println(response.Response().Status)
}

func ipfilter(ip string) bool {
	matchString, err := regexp.MatchString("((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})(\\.((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})){3}", ip)
	if err != nil {
		log.Println("??????IP")
		return false
	}
	if matchString {

		addr, err := net.ResolveIPAddr("ip", "google.com")
		if err != nil || addr.String() == ip {
			log.Println("??? google")
			return false
		}

		addr, err = net.ResolveIPAddr("ip", "baidu.com")
		if err != nil || addr.String() == ip {
			log.Println("??? ??????")
			return false
		}

		addr, err = net.ResolveIPAddr("ip", "www.google.com")
		if err != nil || addr.String() == ip {
			log.Println("??? google")
			return false
		}

		addr, err = net.ResolveIPAddr("ip", "www.baidu.com")
		if err != nil || addr.String() == ip {
			log.Println("??? ??????")
			return false
		}

		ipreg := `^10\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])$`
		matchString, err = regexp.MatchString(ipreg, ip)
		if err != nil || matchString {
			log.Println("??? 10???")
			return false
		}

		ipreg = `^172\.(1[6789]|2[0-9]|3[01])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])$`
		matchString, err = regexp.MatchString(ipreg, ip)
		if err != nil || matchString {
			log.Println("??? 172???")
			return false
		}
		ipreg = `^127\.(1[6789]|2[0-9]|3[01])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])$`
		matchString, err = regexp.MatchString(ipreg, ip)
		if err != nil || matchString {
			log.Println("??? 127???")
			return false
		}

		ipreg = `^192\.168\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[0-9])$`
		matchString, err = regexp.MatchString(ipreg, ip)
		if err != nil || matchString {
			log.Println("??? 192???")
			return false
		}
		return true
	}
	return false
}

func taskpush(m *pushmsg, cache *db.RedisClient) {
	rand.Seed(time.Now().Unix())
	var startipa int
	ipa, err := cache.Get("scanip-a")
	if err != nil {
		startipa = 1
	} else {
		startipa, _ = strconv.Atoi(ipa)
	}

	var startipb int
	ipb, err := cache.Get("scanip-b")
	if err != nil {
		startipb = 0
	} else {
		startipb, _ = strconv.Atoi(ipb)
	}

	var startipc int
	ipc, err := cache.Get("scanip-c")
	if err != nil {
		startipc = 0
	} else {
		startipc, _ = strconv.Atoi(ipc)
	}
	for a := startipa; a < 255; a++ {
		for b := startipb; b < 255; b++ {
			for c := startipc; c < 255; c++ {
				ip := fmt.Sprintf("%v.%v.%v.1/24", a, b, c)
				if !ipfilter(strings.Split(ip, "/")[0]) {
					continue
				}
				logs.Install().Infoln(ip)
				var scner string
				var rate int
				if rand.Int63n(2) == 0 {
					scner = "syn"
					rate = 500
				} else {
					scner = "masscan"
					rate = 5000
				}
				var portlist []interface{}
				config.Install().Reget("ports", &portlist)
				var ports string
				for _, v := range portlist {
					if ports == "" {
						ports = fmt.Sprintf("%v", v)
					} else {
						ports = fmt.Sprintf("%v,%v", ports, v)
					}

				}
				marshal, err := json.Marshal(map[string]interface{}{
					"ip":     ip,
					"scaner": scner,
					"rate":   rate,
					"ports":  ports,
				})
				if err != nil {
					continue
				}
				for getcount("scanport", m.Getqueues) > 100 {
					logs.Install().Infoln("????????????")
					time.Sleep(1 * time.Second)
				}
				push("scanport", string(marshal), m.Pushurl)
				cache.Set("scanip-a", a, time.Hour)
				cache.Set("scanip-b", b, time.Minute)
				cache.Set("scanip-c", c, time.Minute)
				if a >= 254 {
					cache.Set("scanip-a", 0, time.Hour)
				}
			}
		}
	}
}

func test(msg pushmsg) {
	marshal, _ := json.Marshal(map[string]interface{}{
		"ip":     "172.16.10.110",
		"scaner": "masscan",
		"rate":   10000,
	})
	push("scanport", string(marshal), msg.Pushurl)
	marshal, _ = json.Marshal(map[string]interface{}{
		"ip":     "172.16.10.110",
		"scaner": "syn",
		"rate":   1000,
	})
	push("scanport", string(marshal), msg.Pushurl)
}
func main() {
	var msg pushmsg
	config.Install().Get("mq", &msg)

	var cachecfg db.RedisConfig
	config.Install().Get("cache", &cachecfg)
	cache := db.NewRedis(&cachecfg)

	if !istopic("scanport", msg.Getqueues) {
		createtopic("scanport", msg.Getqueues)
	}
	taskpush(&msg, cache)
	c := cron.New() // ??????????????????????????????
	c.AddFunc("*/10 * * * *", func() {
		if getcount("scanport", msg.Getqueues) == 0 {
			taskpush(&msg, cache)
		}
	}) // ???????????????????????????
	c.Start()
	select {}
}
