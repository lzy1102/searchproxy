package main

import (
	"fmt"
	"time"
)

func fib(n int) int {
	if n < 2 {
		return n
	} else {
		return fib(n-2) + fib(n-1)
	}
}
func main() {
	starttime := time.Now().Unix()
	fmt.Println("计算结果", fib(42))
	endtime := time.Now().Unix()
	fmt.Println("耗时", endtime-starttime, "秒")
}
