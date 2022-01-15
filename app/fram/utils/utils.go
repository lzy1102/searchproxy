package utils

import (
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func FatalAssert(err error) {
	if nil != err {
		panic(err)
	}
}

// GetCurrentAbPathByExecutable 获取当前执行程序所在的绝对路径
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
			if strings.TrimSpace(v) != "" {
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

//getIpAll 获取ip段
//最小IP： 223.255.0.1  最大IP： 223.255.127.255
func GetIpAll(minIp, maxIp string) []string {
	ipArr := make([]string, 0)
	minIpaddress := net.ParseIP(minIp)
	maxIpaddress := net.ParseIP(maxIp)
	if minIpaddress == nil || maxIpaddress == nil {
		fmt.Println("ip地址格式不正确")
	} else {
		minIpSplitArr := strings.Split(minIp, ".")
		maxIpSplitArr := strings.Split(maxIp, ".")

		minIP1, _ := strconv.Atoi(minIpSplitArr[0])
		minIP2, _ := strconv.Atoi(minIpSplitArr[1])
		minIP3, _ := strconv.Atoi(minIpSplitArr[2])
		minIP4, _ := strconv.Atoi(minIpSplitArr[3])

		maxIP1, _ := strconv.Atoi(maxIpSplitArr[0])
		maxIP2, _ := strconv.Atoi(maxIpSplitArr[1])
		maxIP3, _ := strconv.Atoi(maxIpSplitArr[2])
		maxIP4, _ := strconv.Atoi(maxIpSplitArr[3])

		if minIP1 <= maxIP1 {
			for i1 := minIP1; i1 <= maxIP1; i1++ {
				minIP1 = i1
				var i2 int
				var maxi2 int
				if minIP1 == maxIP1 { //如果第一个数相等
					i2 = minIP2
					maxi2 = maxIP2
				} else {
					i2 = 0
					maxi2 = 255
				}
				for ii2 := i2; ii2 <= maxi2; ii2++ {
					minIP2 = ii2
					var i3 int
					var maxi3 int
					if minIP1 == maxIP1 && minIP2 == maxIP2 { //如果第一个数相等 并且 第二个数相等
						i3 = minIP3
						maxi3 = maxIP3
					} else {
						i3 = 0
						maxi3 = 255
					}
					for ii3 := i3; ii3 <= maxi3; ii3++ {
						minIP3 = ii3
						var i4 int
						var maxi4 int
						if minIP1 == maxIP1 && minIP2 == maxIP2 && minIP3 == maxIP3 { //如果第一个数相等 并且 第二个数相等 并且 第三个数相等
							i4 = minIP4
							maxi4 = maxIP4
						} else {
							i4 = minIP4
							maxi4 = 255
						}
						for ii4 := i4; ii4 <= maxi4; ii4++ {
							minIP4 = ii4
							newIP := fmt.Sprintf("%v.%v.%v.%v", minIP1, minIP2, minIP3, minIP4)
							ipArr = append(ipArr, newIP)

						}
						minIP4 = 0

					}
					minIP3 = 0

				}
				minIP2 = 0

			}
		}
	}
	return ipArr
}
