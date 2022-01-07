package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/tevino/tcp-shaker"
	"time"
)

func main()  {
	var addr  string
	flag.StringVar(&addr,"addr","127.0.0.1:10808","")
	flag.Parse()
	c := tcp.NewChecker()

	ctx, stopChecker := context.WithCancel(context.Background())
	defer stopChecker()
	go func() {
		if err := c.CheckingLoop(ctx); err != nil {
			fmt.Println("checking loop stopped due to fatal error: ", err)
		}
	}()

	<-c.WaitReady()

	timeout := time.Second * 1
	err := c.CheckAddr(addr, timeout)
	switch err {
	case tcp.ErrTimeout:
		fmt.Println("Connect to Google timed out")
	case nil:
		fmt.Println("Connect to Google succeeded")
	default:
		fmt.Println("Error occurred while connecting: ", err)
	}
}