package main

import (
	"fmt"
	"searchproxy/app/fram/config"
)

func fib(n int) int {
	if n < 2 {
		return n
	} else {
		return fib(n-2) + fib(n-1)
	}
}
func main() {
	var portlist []interface{}
	config.Install().Get("ports", &portlist)
	var ports string
	for _, v := range portlist {
		if ports == "" {
			ports = fmt.Sprintf("%v", v)
		} else {
			ports = fmt.Sprintf("%v,%v", ports, v)
		}

	}
}
