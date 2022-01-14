package main

import "log"

func fib(n int) int {
	if n < 2 {
		return n
	} else {
		return fib(n-2) + fib(n-1)
	}
}
func main() {
	num := 0
	for num > 10 {
		log.Println(num)
		num++
	}
}
