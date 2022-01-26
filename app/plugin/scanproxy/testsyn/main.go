// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// synscan implements a TCP syn scanner on top of pcap.
// It's more complicated than arpscan, since it has to handle sending packets
// outside the local network, requiring some routing and ARP work.
//
// Since this is just an example program, it aims for simplicity over
// performance.  It doesn't handle sending packets very quickly, it scans IPs
// serially instead of in parallel, and uses gopacket.Packet instead of
// gopacket.DecodingLayerParser for packet processing.  We also make use of very
// simple timeout logic with time.Since.
//
// Making it blazingly fast is left as an exercise to the reader.
//go:build linux
// +build linux

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
	"log"
	"math"
	"net"
	"searchproxy/app/fram/utils"
	routing2 "searchproxy/app/plugin/scanproxy/testsyn/routing"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

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

func (r *router) getIface(dstip net.IP) *net.Interface {
	for _, iface := range r.ifaces {
		addrs, _ := iface.Addrs()
		for _, address := range addrs {
			ipNet, _ := address.(*net.IPNet)
			if dstip.String() == ipNet.String() {
				return &iface
			}
		}
	}
	return &net.Interface{}
}

// scanner handles scanning a single IP address.
type scanner struct {
	// iface is the interface to send packets on.
	iface *net.Interface
	// destination, gateway (if applicable), and source IP addresses to use.
	dst, gw, src net.IP

	handle *pcap.Handle

	// opts and buf allow us to easily serialize packets in the send()
	// method.
	opts gopacket.SerializeOptions
	buf  gopacket.SerializeBuffer
}

// newScanner creates a new scanner for a given destination IP address, using
// router to determine how to route packets to that IP.
func newScanner(ip net.IP, router routing.Router) (*scanner, error) {
	s := &scanner{
		dst: ip,
		opts: gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		},
		buf: gopacket.NewSerializeBuffer(),
	}
	iface, gw, src, err := router.Route(ip)
	if err != nil {
		return nil, err
	}
	//log.Printf("scanning ip %v with interface %v, gateway %v, src %v", ip, iface.Name, gw, src)
	s.gw, s.src, s.iface = gw, src, iface

	// Open the handle for reading/writing.
	// Note we could very easily add some BPF filtering here to greatly
	// decrease the number of packets we have to look at when getting back
	// scan results.
	handle, err := pcap.OpenLive(iface.Name, 65536, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	s.handle = handle
	return s, nil
}

// close cleans up the handle.
func (s *scanner) close() {
	s.handle.Close()
}

// getHwAddr is a hacky but effective way to get the destination hardware
// address for our packets.  It does an ARP request for our gateway (if there is
// one) or destination IP (if no gateway is necessary), then waits for an ARP
// reply.  This is pretty slow right now, since it blocks on the ARP
// request/reply.
func (s *scanner) getHwAddr() (net.HardwareAddr, error) {
	start := time.Now()
	arpDst := s.dst
	if s.gw != nil {
		arpDst = s.gw
	}
	// Prepare the layers to send for an ARP request.
	eth := layers.Ethernet{
		SrcMAC:       s.iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(s.iface.HardwareAddr),
		SourceProtAddress: []byte(s.src),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(arpDst),
	}
	// Send a single ARP request packet (we never retry a send, since this
	// is just an example ;)
	if err := s.send(&eth, &arp); err != nil {
		return nil, err
	}
	// Wait 3 seconds for an ARP reply.
	for {
		if time.Since(start) > time.Second*5 {
			return nil, errors.New("timeout getting ARP reply")
		}
		data, _, err := s.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			return nil, err
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
			arp := arpLayer.(*layers.ARP)
			if net.IP(arp.SourceProtAddress).Equal(net.IP(arpDst)) {
				return net.HardwareAddr(arp.SourceHwAddress), nil
			}
		}
	}
}

// scan scans the dst IP address of this scanner.
func (s *scanner) scan(dstport int) (bool, error) {
	// First off, get the MAC address we should be sending packets to.
	hwaddr, err := s.getHwAddr()
	if err != nil {
		return false, err
	}
	// Construct all the network layers we need.
	eth := layers.Ethernet{
		SrcMAC:       s.iface.HardwareAddr,
		DstMAC:       hwaddr,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip4 := layers.IPv4{
		SrcIP:    s.src,
		DstIP:    s.dst,
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}
	tcp := layers.TCP{
		SrcPort: 54321,
		DstPort: 0, // will be incremented during the scan
		SYN:     true,
	}
	tcp.SetNetworkLayerForChecksum(&ip4)

	// Create the flow we expect returning packets to have, so we can check
	// against it and discard useless packets.
	ipFlow := gopacket.NewFlow(layers.EndpointIPv4, s.dst, s.src)
	start := time.Now()
	//for {
	// Send one packet per loop iteration until we've sent packets
	// to all of ports [1, 65535].
	//if tcp.DstPort < 65535 {
	start = time.Now()
	tcp.DstPort = layers.TCPPort(dstport)
	//	tcp.DstPort++
	if err := s.send(&eth, &ip4, &tcp); err != nil {
		//log.Printf("error sending to port %v: %v", tcp.DstPort, err)
		return false, err
	}
	//}
	// Time out 5 seconds after the last packet we sent.
	if time.Since(start) > time.Second*5 {
		//log.Printf("timed out for %v, assuming we've seen all we can", s.dst)
		return false, nil
	}

	// Read in the next packet.
	data, _, err := s.handle.ReadPacketData()
	if err == pcap.NextErrorTimeoutExpired {
		//continue
		return false, err
	} else if err != nil {
		//log.Printf("error reading packet: %v", err)
		//continue
		return false, err
	}
	gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
	if tcp.SYN && tcp.ACK {
		//log.Printf("  port %v open", tcp.SrcPort)
		return true, nil
	}

	//// Parse the packet.  We'd use DecodingLayerParser here if we
	//// wanted to be really fast.
	//packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
	//
	//// Find the packets we care about, and print out logging
	//// information about them.  All others are ignored.
	//if net := packet.NetworkLayer(); net == nil {
	//	// log.Printf("packet has no network layer")
	//} else if net.NetworkFlow() != ipFlow {
	//	// log.Printf("packet does not match our ip src/dst")
	//} else if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer == nil {
	//	// log.Printf("packet has not tcp layer")
	//} else if tcp, ok := tcpLayer.(*layers.TCP); !ok {
	//	// We panic here because this is guaranteed to never
	//	// happen.
	//	//panic("tcp layer is not tcp layer :-/")
	//}  else if tcp.RST {
	//	//log.Printf("  port %v closed", tcp.SrcPort)
	//} else if tcp.SYN && tcp.ACK {
	//	//log.Printf("  port %v open", tcp.SrcPort)
	//	return true, nil
	//}
	//}
	return false, nil
}

// send sends the given layers as a single packet on the network.
func (s *scanner) send(l ...gopacket.SerializableLayer) error {
	if err := gopacket.SerializeLayers(s.buf, s.opts, l...); err != nil {
		return err
	}
	return s.handle.WritePacketData(s.buf.Bytes())
}
func getallip(ip string) []string {
	iplist := make([]string, 0)
	if strings.Contains(ip, "/") {
		tmp := strings.Split(ip, "/")
		atoi, err := strconv.Atoi(tmp[1])
		if err != nil {
			return []string{}
		}
		maxhost := int(math.Pow(float64(2), float64(32-atoi))) - 2
		minip := tmp[0]
		tmpip := strings.Split(tmp[0], ".")
		hostid, err := strconv.Atoi(tmpip[3])
		if err != nil {
			return []string{}
		}
		maxip := fmt.Sprintf("%v.%v.%v.%v", tmpip[0], tmpip[1], tmpip[2], hostid+maxhost)
		fmt.Println(minip, maxip)
		iplist = utils.GetIpAll(minip, maxip)
	} else {
		iplist = append(iplist, ip)
	}
	return iplist
}
func main() {
	var ipstr string
	flag.StringVar(&ipstr, "ip", "127.0.0.1", "")
	flag.Parse()

	defer util.Run()()

	router, err := routing2.New()
	if err != nil {
		return
	}
	if err != nil {
		log.Fatal("routing error:", err)
	}
	ratechan := make(chan interface{}, 50) // 控制任务并发的chan
	datachan := make(chan interface{}, 0)
	for _, i2 := range getallip(ipstr) {
		log.Println(i2)
		ratechan <- struct{}{} // 作用类似于waitgroup.Add(1)
		go func(host string) {
			var ip net.IP
			if ip = net.ParseIP(host); ip == nil {
				return
			} else if ip = ip.To4(); ip == nil {
				return
			}
			s, err := newScanner(ip, router)
			if err != nil {
				log.Printf("unable to create scanner for %v: %v", ip, err)
				return
			}
			status, _ := s.scan(10808)
			defer s.close()
			datachan <- map[string]interface{}{
				"ip":     host,
				"port":   10808,
				"status": status,
			}
			<-ratechan // 执行完毕，释放资源
		}(i2)
	}

	for range getallip(ipstr) {
		tmp := <-datachan
		log.Println(tmp)
	}
}
