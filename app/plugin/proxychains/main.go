package main

import (
	"github.com/imroc/req"
	"log"
)

func main() {
	resp, err := req.Get("https://www.google.com")
	if err != nil {
		return
	}
	log.Println(resp.String())
}
