package main

import (
	"github.com/cheggaaa/pb/v3"
	"time"
)

func synscan(rate int) (result []interface{}) {
	ratechan := make(chan interface{}, rate) // 控制任务并发的chan
	datachan := make(chan interface{}, 0)
	bar := pb.StartNew(65535 * 255)
	for s := 0; s < 255; s++ {
		for p := 0; p < 65535; p++ {
			ratechan <- struct{}{} // 作用类似于waitgroup.Add(1)
			go func(host int, port int) {
				time.Sleep(5 * time.Second)
				bar.Increment()
				<-ratechan // 执行完毕，释放资源
				datachan <- map[string]interface{}{
					"ip":     host,
					"port":   port,
					"status": true,
				}
			}(s, p)
		}
	}
	for s := 0; s < 255; s++ {
		for p := 0; p < 65535; p++ {
			tmp := <-datachan
			if proxystatus, ok := tmp.(map[string]interface{})["status"]; ok && proxystatus.(bool) {
				result = append(result, tmp)
			}
			//bar.Increment()
		}
	}
	bar.Finish()
	return result
}
func main() {
	synscan(100000)
}
