package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

// 2 bytes: hardware type
// 2 bytes: protocol type
// 1 byte : hardware address length
// 1 byte : protocol length
// 2 bytes: operation
// N bytes: source hardware address
// N bytes: source protocol address
// N bytes: target hardware address
// N bytes: target protocol address

type OperationCode uint16

const (
	REQUEST OperationCode = 1
	REPLY                 = 2
)

type Packet struct {
	HdwType       uint16
	ProtocolType  uint16
	HdwLen        uint8
	IPLen         uint8
	OP            OperationCode
	SenderHdwAddr net.HardwareAddr
	SenderIP      net.IP
	TargetHdwAddr net.HardwareAddr
	TargetIP      net.IP
}

var (
	ErrWrongMACAddr  = errors.New("Wrong Hardware Address")
	ErrWrongIPAddr   = errors.New("Wrong IP Address")
	ErrUnExpectedEOF = errors.New("UnExpected EOF")
)

func MakePacket(op OperationCode, srcHdwr net.HardwareAddr,
	srcIP net.IP, dstHdwr net.HardwareAddr, dstIP net.IP) (*Packet, error) {

	if len(srcHdwr) < 6 {
		return nil, ErrWrongMACAddr
	}

	if len(srcIP) != 16 || len(dstIP) != 16 {
		return nil, ErrWrongIPAddr
	}

	return &Packet{
		// 1 for Ethernet (10Mb)
		HdwType: 1,
		// EtherType
		ProtocolType:  0x0800,
		HdwLen:        uint8(len(srcHdwr)),
		IPLen:         uint8(len(srcIP)),
		OP:            op,
		SenderHdwAddr: srcHdwr,
		SenderIP:      srcIP,
		TargetHdwAddr: dstHdwr,
		TargetIP:      dstIP,
	}, nil
}

func (pkt *Packet) Marshal() ([]byte, error) {
	knownSize := 8
	b := make([]byte, uint8(knownSize)+(pkt.IPLen*2)+(pkt.HdwLen*2))

	binary.BigEndian.PutUint16(b[:2], pkt.HdwType)
	binary.BigEndian.PutUint16(b[2:4], pkt.ProtocolType)

	b[4] = pkt.HdwLen
	b[5] = pkt.IPLen

	binary.BigEndian.PutUint16(b[6:8], uint16(pkt.OP))

	hdwLen := int(pkt.HdwLen)
	ipLen := int(pkt.IPLen)

	copy(b[knownSize:hdwLen+knownSize], pkt.SenderHdwAddr)

	knownSize += hdwLen
	copy(b[knownSize:ipLen+knownSize], pkt.SenderIP)

	knownSize += ipLen
	copy(b[knownSize:hdwLen+knownSize], pkt.TargetHdwAddr)

	knownSize += hdwLen
	copy(b[knownSize:knownSize+ipLen], pkt.TargetIP)

	return b, nil

}

// Show only for debugging.
func (pkt *Packet) Show() {
	fmt.Printf("HARDWARE TYPE %v\n", pkt.HdwType)
	fmt.Printf("ProtocolType %v\n", pkt.ProtocolType)
	fmt.Printf("HdwLen %v\n", pkt.HdwLen)
	fmt.Printf("IP length %v\n", pkt.IPLen)
	fmt.Printf("OP %v\n", pkt.OP)
	fmt.Printf("SENDER HDW %v\n", pkt.SenderHdwAddr)
	fmt.Printf("SENDER IP %v\n", pkt.SenderIP)
	fmt.Printf("TARGET HDW %v\n", pkt.TargetHdwAddr)
	fmt.Printf("TARGET IP %v\n", pkt.TargetIP)
}

func (pkt *Packet) Unmarshal(b []byte) error {
	if len(b) < 8 {
		return ErrUnExpectedEOF
	}

	pkt.HdwType = binary.BigEndian.Uint16(b[0:2])
	pkt.ProtocolType = binary.BigEndian.Uint16(b[2:4])

	pkt.HdwLen = b[4]
	hdwLen := int(pkt.HdwLen)
	pkt.IPLen = b[5]
	ipLen := int(pkt.IPLen)

	pkt.OP = OperationCode(binary.BigEndian.Uint16(b[6:8]))

	knownSize := 8

	bufferSize := hdwLen*2 + ipLen*2 + knownSize

	buffer := make([]byte, bufferSize)
	copy(buffer[:hdwLen], b[knownSize:knownSize+hdwLen])
	pkt.SenderHdwAddr = buffer[:hdwLen]

	knownSize += hdwLen

	rest := hdwLen + ipLen

	copy(buffer[hdwLen:ipLen+hdwLen], b[knownSize:knownSize+ipLen])
	pkt.SenderIP = buffer[hdwLen : ipLen+hdwLen]

	knownSize += ipLen

	rest += hdwLen

	copy(buffer[ipLen+hdwLen:hdwLen+hdwLen+ipLen], b[knownSize:knownSize+hdwLen])

	pkt.TargetHdwAddr = buffer[ipLen+hdwLen : hdwLen+ipLen+hdwLen]

	knownSize += hdwLen

	rest += ipLen

	copy(buffer[rest-ipLen:rest], b[knownSize:knownSize+ipLen])

	pkt.TargetIP = buffer[rest-ipLen : rest]

	return nil
}
