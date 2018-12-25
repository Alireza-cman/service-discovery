package main

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"golang.org/x/net/ipv4"
)

const (
	// https://en.wikipedia.org/wiki/User_Datagram_Protocol#Packet_structure
	maxDatagramSize = 66507
)

type peerDiscovery struct {
	received map[string][]byte
	sync.RWMutex
}

func main() {
	p := new(peerDiscovery)
	p.received = make(map[string][]byte)
	//var p sync.RWMutex
	p.RLock()
	address := "239.255.255.250:9999"
	portNum := 9999
	multicastAddressNumbers := []uint8{239, 255, 255, 250}
	//allowSelf := p.settings.AllowSelf
	p.RUnlock()
	localIPs := getLocalIPs()

	// get interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	// log.Println(ifaces)

	// Open up a connection
	c, err := net.ListenPacket("udp4", address)
	if err != nil {
		return
	}
	defer c.Close()

	group := net.IPv4(multicastAddressNumbers[0], multicastAddressNumbers[1], multicastAddressNumbers[2], multicastAddressNumbers[3])
	p2 := ipv4.NewPacketConn(c)
	for i := range ifaces {
		if errJoinGroup := p2.JoinGroup(&ifaces[i], &net.UDPAddr{IP: group, Port: portNum}); errJoinGroup != nil {
			// log.Print(errJoinGroup)
			continue
		}
	}

	// Loop forever reading from the socket
	for {
		buffer := make([]byte, maxDatagramSize)
		n, _, src, errRead := p2.ReadFrom(buffer)
		// log.Println(n, src.String(), err, buffer[:n])
		if errRead != nil {
			err = errRead
			return
		}

		if _, ok := localIPs[strings.Split(src.String(), ":")[0]]; ok {
			continue
		}

		// log.Println(src, hex.Dump(buffer[:n]))

		ip := strings.Split(src.String(), ":")[0]

		p.Lock()
		if _, ok := p.received[ip]; !ok {
			p.received[ip] = buffer[:n]
			fmt.Println(ip)
			fmt.Println(string(p.received[ip]))
		}
		p.Unlock()
		//p.RLock()
		//fmt.Println(buffer[:n])
		// if len(p.received) >= p.settings.Limit && p.settings.Limit > 0 {
		// 	p.RUnlock()
		// 	break
		// }
		//p.RUnlock()
	}
	return
}
func getLocalIPs() (ips map[string]struct{}) {
	ips = make(map[string]struct{})
	ips["localhost"] = struct{}{}
	ips["127.0.0.1"] = struct{}{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, address := range addrs {
		ips[strings.Split(address.String(), "/")[0]] = struct{}{}
	}
	return
}
