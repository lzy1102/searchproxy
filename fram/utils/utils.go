package utils

import (
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func FatalAssert(err error) {
	if nil != err {
		panic(err)
	}
}


// 获取当前执行程序所在的绝对路径
func GetCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		FatalAssert(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

func RemoveRepeatedElement(arr []string) (newArr []string) {
	result := make([]string, 0)
	m := make(map[string]bool) //map的值不重要
	for _, v := range arr {
		if _, ok := m[v]; !ok {
			if strings.TrimSpace(v) !=""{
				result = append(result, v)
				m[v] = true
			}
		}
	}
	return result
}


func Int64ToIp(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func Ip2Int64(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}