package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/tevino/tcp-shaker"
	"log"
	"time"
)

func tcpshaker(ip string, port int) bool {
	c := tcp.NewChecker()
	ctx, stopChecker := context.WithCancel(context.TODO())
	defer stopChecker()
	go func() {
		if err := c.CheckingLoop(ctx); err != nil {
			fmt.Println("checking loop stopped due to fatal error: ", err)
		}
	}()

	<-c.WaitReady()

	timeout := 5 * time.Second
	err := c.CheckAddr(fmt.Sprintf("%v:%v", ip, port), timeout)
	if err == nil {
		return true
	}
	return false
}

func scan(ip string, rate int) (result []interface{}) {
	ratechan := make(chan interface{}, rate) // 控制任务并发的chan
	datachan := make(chan interface{}, 0)
	bar := pb.StartNew(65535)
	for i := 1; i <= 65535; i++ {
		ratechan <- struct{}{} // 作用类似于waitgroup.Add(1)
		bar.Increment()
		go func(host string, port int) {
			portstatus := tcpshaker(host, port)
			<-ratechan // 执行完毕，释放资源
			datachan <- map[string]interface{}{
				"ip":     host,
				"port":   port,
				"status": portstatus,
			}
		}(ip, i)
	}
	for i := 1; i <= 65535; i++ {
		tmp := <-datachan
		if proxystatus, ok := tmp.(map[string]interface{})["status"]; ok && proxystatus.(bool) {
			log.Println(tmp.(map[string]interface{})["port"], "open")
		}
	}
	bar.Finish()
	return result
}

func main() {
	var ip string
	var rate int
	flag.StringVar(&ip, "ip", "127.0.0.1", "target ip")
	flag.IntVar(&rate, "rate", 1000, "thread number")
	flag.Parse()
	scan(ip, rate)
}
