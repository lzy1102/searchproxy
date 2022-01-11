package main

import (
	"github.com/imroc/req"
	"io/ioutil"
	"log"
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

func main() {
	createtopic("hello", "")
	get, err := req.Get("http://proxy:3f0c1304c3865ea6@172.16.30.190:15672/api/queues/%2F/")
	if err != nil {
		panic(err)
	}
	var data interface{}
	err = get.ToJSON(&data)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("out.json", get.Bytes(), 0777)
	log.Println(data)
}
