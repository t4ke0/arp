package client

import (
	"bytes"
	"net"
	"strings"

	"github.com/mdlayher/ethernet"
	"github.com/mdlayher/raw"

	"arp/packet"
)

const protocolARP = 0x0806

//TODO: client need to have src ip addr and src mac addr as fields.

// Client ...
type Client struct {
	ifc  net.Interface
	conn *raw.Conn
}

// New ...
func New(netiface string) (*Client, error) {
	client := &Client{}
	const defInterfaceIdx = 0
	//Default interface is the one that is in index 0
	ifcs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	if netiface == "" {
		client.ifc = ifcs[defInterfaceIdx]
	} else {
		client.ifc = getInterfaceByName(netiface, ifcs)
	}
	conn, err := raw.ListenPacket(&client.ifc, protocolARP, &raw.Config{})
	if err != nil {
		return nil, err
	}
	client.conn = conn
	return client, nil
}

// SendTO ...
func (c *Client) SendTO(pkt *packet.Packet, dst net.HardwareAddr) error {
	data, err := pkt.Marshal()
	if err != nil {
		return err
	}

	frame := &ethernet.Frame{
		Destination: dst,
		Source:      pkt.SenderHdwAddr,
		EtherType:   ethernet.EtherTypeARP,
		Payload:     data,
	}

	frameBin, err := frame.MarshalBinary()
	if err != nil {
		return err
	}

	if c != nil {
		if _, err := c.conn.WriteTo(frameBin, &raw.Addr{dst}); err != nil {
			return err
		}
	}
	return nil

}

// ResolveAddr ...
func (c *Client) ResolveAddr(pkt *packet.Packet, dst net.HardwareAddr) (net.HardwareAddr, error) {
	if err := c.SendTO(pkt, dst); err != nil {
		return nil, err
	}

	buffer := make([]byte, 128)
	eFrame := &ethernet.Frame{}

	_, _, err := c.conn.ReadFrom(buffer)
	if err != nil {
		return nil, err
	}

	if err := eFrame.UnmarshalBinary(buffer); err != nil {
		return nil, err
	}

	if eFrame.EtherType == ethernet.EtherTypeARP {
		respPkt := new(packet.Packet)
		respPkt.Unmarshal(eFrame.Payload)
		if respPkt.OP == packet.REPLY {
			if bytes.Compare(pkt.TargetIP, respPkt.SenderIP) == 0 {
				return respPkt.SenderHdwAddr, nil
			}
		}
	}

	return nil, nil
}

func getInterfaceByName(netiface string, netifaces []net.Interface) (foundInterface net.Interface) {
	for _, i := range netifaces {
		if i.Name == netiface {
			foundInterface = i
			return
		}
	}
	return
}

func GetLocalMacAddr(itface string) (net.HardwareAddr, error) {
	ifcs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	if itface != "" {
		for _, n := range ifcs {
			if n.Name == itface {
				return n.HardwareAddr, nil
			}
		}
	}
	return ifcs[0].HardwareAddr, nil
}

func GetSrcIPAddr(itface string) (net.IP, error) {
	ifcs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	getAddr := func(addrs []net.Addr) net.IP {
		for _, a := range addrs {
			if a.Network() == "ip+net" {
				return net.ParseIP(strings.Split(a.String(), "/")[0])
			}
		}
		return nil
	}

	if itface != "" {
		for _, n := range ifcs {
			if n.Name == itface {
				addrs, err := n.Addrs()
				if err != nil {
					return nil, err
				}
				return getAddr(addrs), nil
			}
		}
	}

	addrs, err := ifcs[0].Addrs()
	if err != nil {
		return nil, err
	}
	return getAddr(addrs), nil
}
