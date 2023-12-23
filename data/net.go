package data

import (
	"bytes"
	"fmt"
	"net"
	"tcpip/constants"
	"tcpip/util"
)

type MacAddress [6]byte

type IPAddress [16]byte

type PacketWithAux [constants.MaxPacketBufferSize]byte

type Packet [constants.MaxPacketSize]byte

type ArpHeaderBytes []byte

type Payload [constants.MaxPayloadSize]byte

type NodeNetworkProperties struct {
	Flags          uint
	ArpTable       *ArpTable
	MacTable       *MacTable
	RoutingTable   *Layer3RouteTable
	IsLbConfigured bool
	LB             IPAddress
}

type IntfNetworkProperties struct {
	MAC              MacAddress
	IsIpConfigured   bool
	IsIpConfigBackup bool
	Vlans            [constants.MaxVlanMembership]uint
	IP               IPAddress
	Mask             rune
	IntfL2Mode       int
}

func (properties *NodeNetworkProperties) InitNodeNetworkProperty() {
	properties.Flags = 0
	properties.IsLbConfigured = false
	properties.ArpTable = &ArpTable{
		ArpEntries: Dll{},
	}
	properties.MacTable = &MacTable{
		MacEntries: Dll{},
	}
	properties.RoutingTable = &Layer3RouteTable{
		Routes: Dll{},
	}
	(&properties.ArpTable.ArpEntries).Init()
	(&properties.MacTable.MacEntries).Init()
	(&properties.RoutingTable.Routes).Init()
	copy(properties.LB[:], bytes.Repeat([]byte{0}, len(properties.LB)))
}

func (properties *IntfNetworkProperties) InitIntfProperty() {
	copy(properties.MAC[:], bytes.Repeat([]byte{0}, len(properties.MAC)))
	properties.IsIpConfigured = false
	copy(properties.IP[:], bytes.Repeat([]byte{0}, len(properties.IP)))
	properties.Mask = 0
}

func (node *Node) SetDeviceType(deviceType uint) bool {
	node.Properties.Flags = uint(util.SetBit(byte(node.Properties.Flags), deviceType))
	return true
}

func (node *Node) SetLbAddress(IP IPAddress) bool {

	if util.IsAllBitsZero(IP[:]) {
		panic("IP address cannot be empty")
	}

	node.Properties.IsLbConfigured = true
	copy(node.Properties.LB[:], IP[:])

	node.Properties.RoutingTable.AddDirectRoute(IP, 32)

	return true
}

func (node *Node) SetIntfIPAddress(intfName string, IP IPAddress, mask rune) bool {
	intf := node.GetNodeIntfByName(intfName)

	if intf == nil {
		panic("Interface not found")
	}

	copy(intf.Properties.IP[:], IP[:])

	intf.Properties.Mask = mask
	intf.Properties.IsIpConfigured = true
	node.Properties.RoutingTable.AddDirectRoute(IP, mask)
	return true
}

func (graph *Graph) Print() {
	fmt.Printf("Topology name: %v\n\n", graph.TopologyName)

	for dllNode := graph.Nodes.Next; dllNode != nil; dllNode = dllNode.Next {
		node := dllNode.DllToNode()
		fmt.Println("	Node name: " + node.NodeName)
		fmt.Printf("	Flags: %v, Port: %v", node.Properties.Flags, node.UDPPortNumber)
		if node.Properties.IsLbConfigured {
			fmt.Printf(", LB Address: %s\n", node.Properties.LB.String())
		} else {
			fmt.Println()
		}

		for _, intf := range node.Interfaces {
			if intf == nil {
				break
			}
			intf.Print()

		}
		fmt.Println()
	}
}

func (node *Node) Print() {
	fmt.Println("	Node name: " + node.NodeName)
	fmt.Printf("	Flags: %v, Port: %v", node.Properties.Flags, node.UDPPortNumber)
	if node.Properties.IsLbConfigured {
		fmt.Printf(", LB Address: %s\n", node.Properties.LB.String())
	} else {
		fmt.Println()
	}
}

func (intf *Interface) Print() {
	link := intf.Link
	neighbourNode := intf.GetNeighbourNode()
	fmt.Printf("		IntfName: %v, ", intf.Name.String())
	if intf.Properties.IsIpConfigured {
		fmt.Printf("IP: %s", intf.Properties.IP.String())
	} else {
		fmt.Printf("IP: nil")
	}
	fmt.Printf(", MAC: %s, ", intf.Properties.MAC.String())
	fmt.Printf("Neighbour node: %v, Cost: %v, Mode: %v\n", neighbourNode.NodeName, link.Cost, intf.Properties.IntfL2Mode)
}

func StringToIPAddress(address string) IPAddress {
	ip := net.ParseIP(address)
	if ip == nil {
		return IPAddress{}
	}
	return IPAddress(net.ParseIP(address))
}

func StringToMACAddress(address string) MacAddress {
	mac, err := net.ParseMAC(address)
	if err != nil {
		return MacAddress{}
	}
	return MacAddress(mac)
}

func (ip IPAddress) String() string {
	return net.IP(ip[:]).String()
}

func (mac MacAddress) String() string {
	return net.HardwareAddr(mac[:]).String()
}

func (intf *Interface) IsTrunkInterfaceVLANEnabled(vlanID uint) bool {
	if intf.Properties.IntfL2Mode != constants.TRUNK {
		panic("Invalid L2 mode for trunk interface")
	}

	for i := 0; i < int(constants.MaxVlanMembership); i++ {
		if intf.Properties.Vlans[i] == vlanID {
			return true
		}
	}
	return false
}

func (ip IPAddress) IsRouteLocalDelivery(node *Node) bool {
	if node.Properties.IsLbConfigured && bytes.Equal(ip[:], node.Properties.LB[:]) {
		return true
	}

	for _, intf := range node.Interfaces {
		if intf == nil {
			return false
		}
		if !intf.Properties.IsIpConfigured {
			continue
		}
		if bytes.Equal(ip[:], intf.Properties.IP[:]) {
			return true

		}
	}
	return false
}
