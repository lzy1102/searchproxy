package main

import (
	"github.com/imroc/req"
	"log"
)

func main() {
	r, err := req.Post("http://proxy:3f0c1304c3865ea6@172.16.30.190:15672/api/exchanges/%2F/amq.default/publish",req.BodyJSON(req.Param{
		"delivery_mode":    "1",
		"headers":          req.Param{},
		"name":             "amq.default",
		"payload":          "fdasfda",
		"payload_encoding": "string",
		"properties":       req.Param{"delivery_mode": 1, "headers": req.Param{}},
		"props":            req.Param{},
		"routing_key":      "test",
		"vhost":            "/",
	}) ,req.Header{"x-vhost":"",
		"User-Agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36",
	} )
	if err!=nil {
		panic(err)
	}
	log.Println(r.String())
}