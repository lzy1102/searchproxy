//+build linux

package main

import (
	"fmt"
	"github.com/google/gopacket/routing"
	"log"
	"net"
	"sort"
	"syscall"
	"unsafe"
)

// rtInfo contains information on a single route.
type rtInfo struct {
	Src, Dst                *net.IPNet
	Gateway, PrefSrc        net.IP
	InputIface, OutputIface uint32
	Priority                uint32
}

// routeSlice implements sort.Interface to sort routes by Priority.
type routeSlice []*rtInfo
type ipAddrs struct {
	v4, v6 net.IP
}
type router struct {
	ifaces []net.Interface
	addrs  []ipAddrs
	v4, v6 routeSlice
}

func New() (routing.Router, error) {
	rtr := &router{}
	tab, err := syscall.NetlinkRIB(syscall.RTM_GETROUTE, syscall.AF_UNSPEC)
	if err != nil {
		return nil, err
	}
	msgs, err := syscall.ParseNetlinkMessage(tab)
	if err != nil {
		return nil, err
	}
loop:
	for _, m := range msgs {
		switch m.Header.Type {
		case syscall.NLMSG_DONE:
			break loop
		case syscall.RTM_NEWROUTE:
			rt := (*routeInfoInMemory)(unsafe.Pointer(&m.Data[0]))
			routeInfo := rtInfo{}
			attrs, err := syscall.ParseNetlinkRouteAttr(&m)
			if err != nil {
				return nil, err
			}
			switch rt.Family {
			case syscall.AF_INET:
				rtr.v4 = append(rtr.v4, &routeInfo)
			case syscall.AF_INET6:
				rtr.v6 = append(rtr.v6, &routeInfo)
			default:
				continue loop
			}
			for _, attr := range attrs {
				switch attr.Attr.Type {
				case syscall.RTA_DST:
					routeInfo.Dst = &net.IPNet{
						IP:   net.IP(attr.Value),
						Mask: net.CIDRMask(int(rt.DstLen), len(attr.Value)*8),
					}
				case syscall.RTA_SRC:
					routeInfo.Src = &net.IPNet{
						IP:   net.IP(attr.Value),
						Mask: net.CIDRMask(int(rt.SrcLen), len(attr.Value)*8),
					}
				case syscall.RTA_GATEWAY:
					routeInfo.Gateway = net.IP(attr.Value)
				case syscall.RTA_PREFSRC:
					routeInfo.PrefSrc = net.IP(attr.Value)
				case syscall.RTA_IIF:
					routeInfo.InputIface = *(*uint32)(unsafe.Pointer(&attr.Value[0]))
				case syscall.RTA_OIF:
					routeInfo.OutputIface = *(*uint32)(unsafe.Pointer(&attr.Value[0]))
				case syscall.RTA_PRIORITY:
					routeInfo.Priority = *(*uint32)(unsafe.Pointer(&attr.Value[0]))
				}
			}
		}
	}
	sort.Sort(rtr.v4)
	sort.Sort(rtr.v6)
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for i, iface := range ifaces {
		if i != iface.Index-1 {
			return nil, fmt.Errorf("out of order iface %d = %v", i, iface)
		}
		rtr.ifaces = append(rtr.ifaces, iface)
		var addrs ipAddrs
		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range ifaceAddrs {
			if inet, ok := addr.(*net.IPNet); ok {
				// Go has a nasty habit of giving you IPv4s as ::ffff:1.2.3.4 instead of 1.2.3.4.
				// We want to use mapped v4 addresses as v4 preferred addresses, never as v6
				// preferred addresses.
				if v4 := inet.IP.To4(); v4 != nil {
					if addrs.v4 == nil {
						addrs.v4 = v4
					}
				} else if addrs.v6 == nil {
					addrs.v6 = inet.IP
				}
			}
		}
		rtr.addrs = append(rtr.addrs, addrs)
	}
	return rtr, nil
}
func main() {
	r,err:= New()
	if err!=nil {
		log.Println(err)
	}
	log.Println(r)
}
