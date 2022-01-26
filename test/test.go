package main

import (
	"encoding/json"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"log"
	"net"
	"sort"
	"syscall"
	"time"
	"unsafe"
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

type rtInfo struct {
	Dst              net.IPNet
	Gateway, PrefSrc net.IP
	OutputIface      uint32
	Priority         uint32
}

type routeSlice []*rtInfo
type router struct {
	ifaces []net.Interface
	addrs  []net.IP
	v4     routeSlice
}

func getRouteInfo() (*router, error) {
	rtr := &router{}
	tab, err := syscall.NetlinkRIB(syscall.RTM_GETROUTE, syscall.AF_INET)
	if err != nil {
		return nil, err
	}
	msgs, err := syscall.ParseNetlinkMessage(tab)
	if err != nil {
		return nil, err
	}
	for _, m := range msgs {
		switch m.Header.Type {
		case syscall.NLMSG_DONE:
			break
		case syscall.RTM_NEWROUTE:
			rtmsg := (*syscall.RtMsg)(unsafe.Pointer(&m.Data[0]))
			attrs, err := syscall.ParseNetlinkRouteAttr(&m)
			if err != nil {
				return nil, err
			}
			routeInfo := rtInfo{}
			rtr.v4 = append(rtr.v4, &routeInfo)
			for _, attr := range attrs {
				switch attr.Attr.Type {
				case syscall.RTA_DST:
					routeInfo.Dst.IP = net.IP(attr.Value)
					routeInfo.Dst.Mask = net.CIDRMask(int(rtmsg.Dst_len), len(attr.Value)*8)
					log.Println("dst.ip", routeInfo.Dst.IP)
				case syscall.RTA_GATEWAY:
					routeInfo.Gateway = net.IPv4(attr.Value[0], attr.Value[1], attr.Value[2], attr.Value[3])
					log.Println("gateway", routeInfo.Gateway)
				case syscall.RTA_OIF:
					routeInfo.OutputIface = *(*uint32)(unsafe.Pointer(&attr.Value[0]))
				case syscall.RTA_PRIORITY:
					routeInfo.Priority = *(*uint32)(unsafe.Pointer(&attr.Value[0]))
				case syscall.RTA_PREFSRC:
					routeInfo.PrefSrc = net.IPv4(attr.Value[0], attr.Value[1], attr.Value[2], attr.Value[3])
					log.Println("prefsrc", routeInfo.PrefSrc)
				}
			}
		}
	}

	sort.Slice(rtr.v4, func(i, j int) bool {
		return rtr.v4[i].Priority < rtr.v4[j].Priority
	})

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for i, iface := range ifaces {

		if i != iface.Index-1 {
			break
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		rtr.ifaces = append(rtr.ifaces, iface)
		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		var addrs net.IP
		for _, addr := range ifaceAddrs {
			if inet, ok := addr.(*net.IPNet); ok {
				if v4 := inet.IP.To4(); v4 != nil {
					if addrs == nil {
						addrs = v4
					}
				}
			}
		}
		rtr.addrs = append(rtr.addrs, addrs)
	}
	return rtr, nil
}
func (r *router) getRoute(dst net.IP) {
	for _, rt := range r.v4 {
		if rt.Dst.IP != nil && !rt.Dst.Contains(dst) {
			continue
		}
		fmt.Printf("%-15v : ", dst.String())
		if rt.PrefSrc == nil {
			fmt.Println(r.ifaces[rt.OutputIface-1].Name, rt.Gateway.String(), r.addrs[rt.OutputIface-1].String())
		} else {
			fmt.Println(r.ifaces[rt.OutputIface-1].Name, rt.Gateway.String(), rt.PrefSrc.String())
		}
		return
	}
}
func main() {

	newRoute, err := getRouteInfo()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("**************************************")
	fmt.Printf("%-15v %-15v %-15v\n", "interfaceName", "gateway", "ip")
	for _, rt := range newRoute.v4 {
		if rt.Gateway != nil {
			log.Println(rt.Gateway, rt.PrefSrc, newRoute.addrs)
			log.Println(json.Marshal(newRoute))
			//fmt.Printf("%-15v %-15v %-15v\n", newRoute.ifaces[rt.OutputIface-1].Name, rt.Gateway.String(), newRoute.addrs[rt.OutputIface-1])
		}
	}
	fmt.Println("**************************************")

	//newRoute.getRoute(net.ParseIP(os.Args[1]))

}
