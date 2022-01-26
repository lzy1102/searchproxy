//go:build linux
// +build linux

package testrouting

import (
	"errors"
	"fmt"
	"github.com/google/gopacket/routing"
	"log"
	"net"
	"sort"
	"syscall"
	"unsafe"
)

// Pulled from http://man7.org/linux/man-pages/man7/rtnetlink.7.html
// See the section on RTM_NEWROUTE, specifically 'struct rtmsg'.
type routeInfoInMemory struct {
	Family byte
	DstLen byte
	SrcLen byte
	TOS    byte

	Table    byte
	Protocol byte
	Scope    byte
	Type     byte

	Flags uint32
}

// rtInfo contains information on a single route.
type rtInfo struct {
	Src, Dst                *net.IPNet
	Gateway, PrefSrc        net.IP
	InputIface, OutputIface uint32
	Priority                uint32
}

// routeSlice implements sort.Interface to sort routes by Priority.
type routeSlice []*rtInfo

func (r routeSlice) Len() int {
	return len(r)
}
func (r routeSlice) Less(i, j int) bool {
	return r[i].Priority < r[j].Priority
}
func (r routeSlice) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

type ipAddrs struct {
	v4, v6 net.IP
}
type router struct {
	ifaces []net.Interface
	addrs  []ipAddrs
	v4, v6 routeSlice
}

func (r *router) Route(dst net.IP) (iface *net.Interface, gateway, preferredSrc net.IP, err error) {
	return r.RouteWithSrc(nil, nil, dst)
}
func (r *router) localIPPort(dstip net.IP) net.IP {
	serverAddr, err := net.ResolveUDPAddr("udp", dstip.String()+":12345")
	if err != nil {
		log.Fatal(err)
	}

	// We don't actually connect to anything, but we can determine
	// based on our destination ip what source ip we should use.
	if con, err := net.DialUDP("udp", nil, serverAddr); err == nil {
		if udpaddr, ok := con.LocalAddr().(*net.UDPAddr); ok {
			return udpaddr.IP
		}
	}
	log.Fatal("could not get local ip: " + err.Error())
	return nil
}
func (r *router) RouteWithSrc(input net.HardwareAddr, src, dst net.IP) (iface *net.Interface, gateway, preferredSrc net.IP, err error) {
	//var ifaceIndex int
	log.Println(r.v4, input, src, dst)
	switch {
	case dst.To4() != nil:
		iface, gateway, preferredSrc, err = r.route(r.v4, input, src, dst)
	case dst.To16() != nil:
		iface, gateway, preferredSrc, err = r.route(r.v6, input, src, dst)
	default:
		err = errors.New("IP is not valid as IPv4 or IPv6")
	}

	if err != nil {
		return
	}

	// Interfaces are 1-indexed, but we store them in a 0-indexed array.
	//ifaceIndex++
	//
	//iface = &r.ifaces[ifaceIndex]
	if preferredSrc == nil {
		preferredSrc = r.localIPPort(dst)
		//switch {
		//case dst.To4() != nil:
		//	//preferredSrc = r.addrs[ifaceIndex].v4
		//
		//	preferredSrc=r.localIPPort(dst)
		//case dst.To16() != nil:
		//	//preferredSrc = r.addrs[ifaceIndex].v6
		//	preferredSrc=r.localIPPort(dst)
		//}
	}
	return
}

func (r *router) route(routes routeSlice, input net.HardwareAddr, src, dst net.IP) (iface *net.Interface, gateway, preferredSrc net.IP, err error) {
	var inputIndex uint32
	//if input != nil {
	for _, ifa := range r.ifaces {
		//if bytes.Equal(input, ifa.HardwareAddr) {
		//	// Convert from zero- to one-indexed.
		//	inputIndex = uint32(i + 1)
		//	break
		//}
		addrs, _ := ifa.Addrs()
		for _, address := range addrs {
			ipNet, _ := address.(*net.IPNet)
			log.Println(ipNet.String(), ifa.Name)
			if r.localIPPort(dst).String() == ipNet.String() {
				iface = &ifa
			}
		}
	}
	//}
	for _, rt := range routes {
		if rt.InputIface != 0 && rt.InputIface != inputIndex {
			continue
		}
		if rt.Src != nil && !rt.Src.Contains(src) {
			continue
		}
		if rt.Dst != nil && !rt.Dst.Contains(dst) {
			continue
		}
		return iface, rt.Gateway, rt.PrefSrc, nil
	}
	err = fmt.Errorf("no route found for %v", dst)
	return
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
		//if i != iface.Index-1 {
		//	return nil, fmt.Errorf("out of order iface %d = %v", i, iface)
		//}
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
		log.Println(i, iface.Index-1, addrs)
		rtr.addrs = append(rtr.addrs, addrs)
	}
	return rtr, nil
}
func main() {
	r, err := New()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(r)
}
