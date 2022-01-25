//go:build linux
// +build linux

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/dean2021/go-masscan"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/tevino/tcp-shaker"
	"io/ioutil"
	"log"
	"math"
	"net"
	"searchproxy/app/fram/utils"
	"strconv"
	"strings"
	"time"
)

// get the local ip and port based on our destination ip
func localIPPort(dstip net.IP) (net.IP, int) {
	serverAddr, err := net.ResolveUDPAddr("udp", dstip.String()+":12345")
	if err != nil {
		log.Fatal(err)
	}

	// We don't actually connect to anything, but we can determine
	// based on our destination ip what source ip we should use.
	if con, err := net.DialUDP("udp", nil, serverAddr); err == nil {
		if udpaddr, ok := con.LocalAddr().(*net.UDPAddr); ok {
			return udpaddr.IP, udpaddr.Port
		}
	}
	log.Fatal("could not get local ip: " + err.Error())
	return nil, -1
}

func socketsyn(host string, port int) bool {

	dstaddrs, err := net.LookupIP(host)
	if err != nil {
		log.Fatal(err)
	}
	// parse the destination host and port from the command line os.Args
	dstip := dstaddrs[0].To4()
	//var dstport layers.TCPPort
	dstport := layers.TCPPort(port)

	srcip, sport := localIPPort(dstip)
	srcport := layers.TCPPort(sport)
	//log.Printf("using srcip: %v", srcip.String())

	// Our IP header... not used, but necessary for TCP checksumming.
	ip := &layers.IPv4{
		SrcIP:    srcip,
		DstIP:    dstip,
		Protocol: layers.IPProtocolTCP,
	}
	// Our TCP header
	tcp := &layers.TCP{
		SrcPort: srcport,
		DstPort: dstport,
		Seq:     1105024978,
		SYN:     true,
		Window:  14600,
	}
	_ = tcp.SetNetworkLayerForChecksum(ip)

	// Serialize.  Note:  we only serialize the TCP layer, because the
	// socket we get with net.ListenPacket wraps our data in IPv4 packets
	// already.  We do still need the IP layer to compute checksums
	// correctly, though.
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	if err := gopacket.SerializeLayers(buf, opts, tcp); err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenPacket("ip4:tcp", "0.0.0.0")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	//log.Println("writing request")
	if _, err := conn.WriteTo(buf.Bytes(), &net.IPAddr{IP: dstip}); err != nil {
		log.Fatal(err)
	}

	// Set deadline so we don't wait forever.
	if err := conn.SetDeadline(time.Now().Add(time.Duration(myinfo.timeout) * time.Second)); err != nil {
		log.Fatal(err)
	}

	for {
		b := make([]byte, 4096)
		//log.Println("reading from conn")
		n, addr, err := conn.ReadFrom(b)
		if err != nil {
			//log.Println("error reading packet: ", err)
			return false
		} else if addr.String() == dstip.String() {
			// Decode a packet
			packet := gopacket.NewPacket(b[:n], layers.LayerTypeTCP, gopacket.Default)
			// Get the TCP layer from this packet
			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				tcp, _ := tcpLayer.(*layers.TCP)

				if tcp.DstPort == srcport {
					if tcp.SYN && tcp.ACK {
						fmt.Println(host, "Port", dstport, " is OPEN")
						return true
					} else {
						//log.Printf("Port %d is CLOSED\n", dstport)
						return false
					}
					return false
				}
			}
			return false
		} else {
			log.Printf("Got packet not matching addr")
			return false
		}
	}
}

func tcpshaker(ip string, port int) bool {
	c := tcp.NewChecker()
	ctx, stopChecker := context.WithCancel(context.TODO())
	defer stopChecker()
	go func() {
		if err := c.CheckingLoop(ctx); err != nil {
			fmt.Println("checking loop stopped due to fatal error: ", err)
		}
	}()

	<-c.WaitReady()

	timeout := time.Duration(myinfo.timeout) * time.Second
	err := c.CheckAddr(fmt.Sprintf("%v:%v", ip, port), timeout)
	if err == nil {
		return true
	}
	return false
}

func synscan(ip, ports string, rate int) (result []interface{}) {
	ratechan := make(chan interface{}, rate) // 控制任务并发的chan
	datachan := make(chan interface{}, 0)
	iplist := getallip(ip)
	var portlist []int
	for _, v := range strings.Split(strings.TrimSpace(ports), ",") {
		if strings.Contains(strings.TrimSpace(v), "-") {
			tmp := strings.Split(v, "-")
			startport, err := strconv.Atoi(tmp[0])
			endport, err := strconv.Atoi(tmp[len(tmp)-1])
			if err != nil {
				continue
			}
			for i := startport; i <= endport; i++ {
				portlist = append(portlist, i)
			}
		} else {
			atoi, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			portlist = append(portlist, atoi)
		}
	}
	bar := pb.StartNew(len(portlist) * len(iplist))
	for _, s := range iplist {
		for _, p := range portlist {
			ratechan <- struct{}{} // 作用类似于waitgroup.Add(1)

			go func(host string, port int) {
				//portstatus := tcpshaker(host, port)
				portstatus := socketsyn(host, port)
				if portstatus {
					log.Println(host, port, "status", portstatus)
				}
				bar.Increment()
				<-ratechan // 执行完毕，释放资源
				datachan <- map[string]interface{}{
					"ip":     host,
					"port":   port,
					"status": portstatus,
				}
			}(s, p)
		}
	}
	for range iplist {
		for range portlist {
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

func masscaner(target, ports, rate string) []interface{} {
	m := masscan.New()
	// masscan可执行文件路径,默认不需要设置
	//m.SetSystemPath("D:\\Program Files\\masscan/masscan.exe")
	// 扫描端口范围
	m.SetPorts(ports)
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
					"ip":     result.Address.Addr,
					"port":   port.Portid,
					"status": true,
				})
			}
		}
	}

	return data
}

type info struct {
	ip      string
	rate    int
	scaner  string
	ports   string
	timeout int
	out     string
}

var myinfo info

func init() {
	flag.StringVar(&myinfo.ip, "ip", "127.0.0.1", "target ip")
	flag.IntVar(&myinfo.rate, "rate", 1000, "thread number")
	flag.StringVar(&myinfo.scaner, "scaner", "syn", "scan name")
	flag.StringVar(&myinfo.ports, "ports", "1080", "port list")
	flag.IntVar(&myinfo.timeout, "timeout", 3, "time out sec")
	flag.StringVar(&myinfo.out, "out", "out.json", "out json file name")
	flag.Parse()
	if myinfo.ip == "" {
		log.Fatalln("ip is err")
	}
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
	var result []interface{}

	if myinfo.scaner == "syn" {
		result = synscan(myinfo.ip, myinfo.ports, myinfo.rate)
	} else if myinfo.scaner == "masscan" {
		result = masscaner(myinfo.ip, myinfo.ports, fmt.Sprintf("%v", myinfo.rate))
	}

	if result != nil && len(result) > 0 {
		out, err := json.Marshal(result)
		if err == nil {
			_ = ioutil.WriteFile(myinfo.out, out, 0777)
		}
	}
}
