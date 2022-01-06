package main

import (
	"github.com/google/gopacket/routing"
	"log"
)

func main() {
	route, err := routing.New()
	if err!=nil {
		log.Println(err)
	}
	log.Println(route)
}
