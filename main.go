package main

import (
	"fmt"
	"log"
	"net"

	"arp/client"
	"arp/packet"
)

func main() {
	c, err := client.New("enp0s25")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(c)

	srcIP, err := client.GetSrcIPAddr("enp0s25")
	if err != nil {
		log.Fatal(err)
	}
	srcHdwr, err := client.GetLocalMacAddr("enp0s25")
	if err != nil {
		log.Fatal(err)
	}

	dstIP := net.ParseIP("10.0.0.1")

	broadCastHdw := net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	pkt, err := packet.MakePacket(packet.REQUEST, srcHdwr, srcIP, broadCastHdw, dstIP)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(pkt)

	data, err := pkt.Marshal()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(data)

	nP := new(packet.Packet)

	if err := nP.Unmarshal(data); err != nil {
		log.Fatal(err)
	}

	nP.Show()
}
