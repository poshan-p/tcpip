package data

import (
	"bytes"
	"crypto/rand"
	"net"
	"tcpip/constants"
	"unsafe"
)

type Graph struct {
	TopologyName string
	Nodes        Dll
}

type Node struct {
	NodeName      string
	Interfaces    [constants.MaxIntfPerNode]*Interface
	Properties    NodeNetworkProperties
	UDPPortNumber uint
	UDPConn       *net.UDPConn
	GraphGlue     Dll
}

type InterfaceName [16]byte

type Interface struct {
	Name       InterfaceName
	Node       *Node
	Link       *Link
	Properties IntfNetworkProperties
}

type Link struct {
	Interface1 Interface
	Interface2 Interface
	Cost       uint
}

func (dll *Dll) DllToNode() *Node {
	return (*Node)(unsafe.Pointer(uintptr(unsafe.Pointer(dll)) - unsafe.Offsetof(Node{}.GraphGlue)))
}

func CreateGraph(topologyName string) *Graph {
	graph := &Graph{
		TopologyName: topologyName,
		Nodes:        Dll{},
	}
	(&graph.Nodes).Init()
	return graph
}

func (graph *Graph) CreateNode(nodeName string) *Node {
	node := &Node{
		NodeName: nodeName,
	}
	(&node.GraphGlue).Init()
	(&node.Properties).InitNodeNetworkProperty()
	(&graph.Nodes).AddNode(&node.GraphGlue)
	return node
}

func (graph *Graph) GetNodeByName(nodeName string) *Node {
	for dllNode := graph.Nodes.Next; dllNode != nil; dllNode = dllNode.Next {
		node := dllNode.DllToNode()
		if node.NodeName == nodeName {
			return node
		}
	}
	return nil
}

func (intf *Interface) GetNeighbourNode() *Node {
	if intf.Node == nil || intf.Link == nil {
		panic("Invalid interface: attached node or link is nil")
	}

	link := intf.Link

	if &link.Interface1 == intf {
		return link.Interface2.Node
	}
	return link.Interface1.Node
}

func (node *Node) GetNodeIntfByName(intfName string) *Interface {
	var intf *Interface

	for i := 0; i < constants.MaxIntfPerNode; i++ {
		intf = node.Interfaces[i]
		if intf == nil {
			return nil
		}
		name := intf.Name.String()

		if name == intfName {
			return intf
		}
	}
	return nil
}

func (node *Node) GetNodeIntfAvailableSlot() int {
	for i := 0; i < constants.MaxIntfPerNode; i++ {
		if node.Interfaces[i] != nil {
			continue
		}
		return i
	}
	return -1
}

func InsertLink(node1, node2 *Node, fromName, toName string, cost uint) {
	link := &Link{
		Interface1: Interface{
			Name: StringToInterfaceName(fromName),
			Node: node1,
		},
		Interface2: Interface{
			Name: StringToInterfaceName(toName),
			Node: node2,
		},
		Cost: cost,
	}
	link.Interface1.Link = link
	link.Interface2.Link = link

	(&link.Interface1).InterfaceAssignMACAddress()
	(&link.Interface2).InterfaceAssignMACAddress()

	var emptyIntfSlot = node1.GetNodeIntfAvailableSlot()
	node1.Interfaces[emptyIntfSlot] = &link.Interface1

	emptyIntfSlot = node2.GetNodeIntfAvailableSlot()
	node2.Interfaces[emptyIntfSlot] = &link.Interface2
}

func (intf *Interface) InterfaceAssignMACAddress() {
	prefix := intf.Name.String() + intf.Node.NodeName

	randomBytes := make([]byte, 3)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return
	}

	copy(intf.Properties.MAC[:3], prefix)
	copy(intf.Properties.MAC[3:], randomBytes)

	intf.Properties.MAC[0] |= 0x02
}

func (node *Node) GetMatchingSubnetInterface(IP IPAddress) *Interface {
	for _, intf := range node.Interfaces {
		if intf == nil || !intf.Properties.IsIpConfigured {
			continue
		}

		subnet1 := applyMask(IP, intf.Properties.Mask)
		subnet2 := applyMask(intf.Properties.IP, intf.Properties.Mask)

		if bytes.Equal(subnet1[:], subnet2[:]) {
			return intf
		}
	}
	return nil
}

func applyMask(ip IPAddress, mask rune) IPAddress {
	return IPAddress(net.IP(ip[:]).Mask(net.CIDRMask(int(mask), 32)).To16())
}

func (interfaceName InterfaceName) String() string {
	return string(interfaceName[:bytes.IndexByte(interfaceName[:], 0)])
}

func StringToInterfaceName(str string) [16]byte {
	var name [16]byte
	copy(name[:], str)
	return name
}
