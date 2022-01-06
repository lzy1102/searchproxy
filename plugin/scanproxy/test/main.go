package main

import (
	"github.com/google/gopacket/routing"
	"log"
)

func main()  {
	router, err := routing.New()
	if err!=nil {
		log.Fatal(err)
	}
	log.Println(router)
}