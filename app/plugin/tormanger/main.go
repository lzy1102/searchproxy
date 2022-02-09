package main

import (
	"fmt"
	"io/ioutil"
)

func settorrc(protocol, ip, port string) {
	if protocol == "socks5" {
		ioutil.WriteFile("/etc/tor/torrc", []byte(fmt.Sprintf("Socks5Proxy %v:%v", ip, port)), 777)
	}

}

func main() {

}
