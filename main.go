package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"arp/client"
	"arp/packet"
)

func getHdwAddrByIP(ip, itfc string) (string, error) {
	c, err := client.New(itfc)
	if err != nil {
		return "", err
	}

	//TODO: steps for getting src ip and src hdwr addr need to be in `client.New` function
	srcIP, err := client.GetSrcIPAddr(itfc)
	if err != nil {
		return "", err
	}

	srcHdwr, err := client.GetLocalMacAddr(itfc)
	if err != nil {
		return "", err
	}

	dstIP := net.ParseIP(ip).To4()

	pkt, err := packet.MakePacket(packet.REQUEST, srcHdwr, srcIP.To4(), packet.Brodcast, dstIP)
	if err != nil {
		return "", err
	}

	addr, err := c.ResolveAddr(pkt, packet.Brodcast)
	if err != nil {
		return "", err
	}

	return addr.String(), nil

}

func main() {

	itfc := flag.String("interface", "", "network interface")
	internetProtocol := flag.String("ip", "", "destination ip")
	flag.Parse()
	if *itfc == "" || *internetProtocol == "" {
		flag.PrintDefaults()
		return
	}

	addr, err := getHdwAddrByIP(*internetProtocol, *itfc)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`
TARGET IP: %s
TARGET MAC: %s
`, *internetProtocol, addr))

}
