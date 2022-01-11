package main

import (
	"fmt"
	"github.com/dean2021/go-masscan"
	"log"
)

func masscaner(target, rate string) []interface{} {
	m := masscan.New()
	// masscan可执行文件路径,默认不需要设置
	//m.SetSystemPath("D:\\Program Files\\masscan/masscan.exe")
	// 扫描端口范围
	m.SetPorts("0-65535")
	// 扫描IP范围
	m.SetRanges(target)
	// 扫描速率
	m.SetRate(rate)
	// 隔离扫描名单
	//m.SetExclude("127.0.0.1")
	// 开始扫描
	err := m.Run()
	if err != nil {
		fmt.Println("scanner failed:", err)
		return nil
	}
	// 解析扫描结果
	results, err := m.Parse()
	if err != nil {
		fmt.Println("Parse scanner result:", err)
		return nil
	}
	data := make([]interface{}, 0)
	for _, result := range results {
		for _, port := range result.Ports {
			if port.State.State == "open" {
				data = append(data, map[string]interface{}{
					"ip":   result.Address.Addr,
					"port": port.Portid,
				})
			}
		}
	}
	return data
}

func main() {
	result := masscaner("45.134.168.29", "1000")
	log.Println(result)
}
