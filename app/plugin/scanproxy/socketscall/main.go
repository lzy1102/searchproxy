package main

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"net"
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
	log.Printf("using srcip: %v", srcip.String())

	// Our IP header... not used, but necessary for TCP checksumming.
	ip := &layers.IPv4{
		SrcIP:    srcip,
		DstIP:    dstip,
		Protocol: layers.IPProtocolTCP,
	}
	// Our TCP header
	t := &layers.TCP{
		SrcPort: srcport,
		DstPort: dstport,
		Seq:     1105024978,
		SYN:     true,
		Window:  14600,
	}
	_ = t.SetNetworkLayerForChecksum(ip)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	if err := gopacket.SerializeLayers(buf, opts, t); err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenPacket("ip4:t", "0.0.0.0")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.WriteTo(buf.Bytes(), &net.IPAddr{IP: dstip}); err != nil {
		log.Fatal(err)
	}
	if err := conn.SetDeadline(time.Now().Add(time.Duration(5) * time.Second)); err != nil {
		log.Fatal(err)
	}

	for {
		b := make([]byte, 4096)
		log.Println("reading from conn")
		n, addr, err := conn.ReadFrom(b)
		if err != nil {
			log.Println("error reading packet: ", err)
			return false
		} else if addr.String() == dstip.String() {
			// Decode a packet
			packet := gopacket.NewPacket(b[:n], layers.LayerTypeTCP, gopacket.Default)
			// Get the TCP layer from this packet
			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				l, _ := tcpLayer.(*layers.TCP)
				if l.DstPort == srcport {
					if l.SYN && l.ACK {
						log.Printf("Port %d is OPEN\n", dstport)
						return true
					} else {
						log.Printf("Port %d is CLOSED\n", dstport)
						return false
					}
					return false
				}
			}
		} else {
			log.Printf("Got packet not matching addr")
		}
	}
}

func main() {
	openport := socketsyn("172.16.10.110", 3389)
	log.Println(openport)
}
