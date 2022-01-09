package main

import (
	"encoding/json"
	"github.com/imroc/req"
	"searchproxy/fram/config"
	"searchproxy/fram/utils"
)

type pushmsg struct {
	pushurl string
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

func main() {
	var msg pushmsg
	config.Install().Get("mq", &msg)
	for i := int64(0); i < utils.Ip2Int64("255.255.255.255"); i++ {
		marshal, err := json.Marshal(map[string]interface{}{
			"ip":   utils.Int64ToIp(i),
			"rate": 1000,
		})
		if err != nil {
			return
		}
		push("scanproxy", string(marshal), msg.pushurl)
	}
}
