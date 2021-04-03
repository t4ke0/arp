package main

import (
	"flag"
	"log"
	"net"

	"arp/client"
	"arp/packet"
)

//NOTE: This where we test the library

func main() {

	itfc := flag.String("interface", "", "network interface")

	internetProtocol := flag.String("ip", "", "destination ip")

	flag.Parse()

	if *itfc == "" || *internetProtocol == "" {
		flag.PrintDefaults()
		return
	}

	c, err := client.New(*itfc)
	if err != nil {
		log.Fatal(err)
	}

	srcIP, err := client.GetSrcIPAddr(*itfc)
	if err != nil {
		log.Fatal(err)
	}

	srcHdwr, err := client.GetLocalMacAddr(*itfc)
	if err != nil {
		log.Fatal(err)
	}

	dstIP := net.ParseIP(*internetProtocol).To4()

	pkt, err := packet.MakePacket(packet.REQUEST, srcHdwr, srcIP.To4(), packet.Brodcast, dstIP)
	if err != nil {
		log.Fatal("ERROR make packet", err)
	}

	if err := c.SendTO(pkt, packet.Brodcast); err != nil {
		log.Fatalf("ERROR sending arp packet %v", err)
	}

	log.Printf("ARP PACKET sent\n")
}
