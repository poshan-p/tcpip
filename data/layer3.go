package data

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"tcpip/constants"
	"unsafe"
)

type IPHeader struct {
	Version       uint8
	IHL           uint8
	TOS           uint8
	Length        uint16
	ID            uint16
	Flags         uint8
	FragmentOffs  uint16
	TTL           uint8
	Protocol      uint8
	Checksum      uint16
	SourceIP      IPAddress
	DestinationIP IPAddress
	Options       []byte
}

type Layer3Route struct {
	DestinationIP IPAddress
	Mask          rune
	IsDirect      bool
	GatewayIP     IPAddress
	InterfaceName InterfaceName
	RouteGlue     Dll
}

type Layer3RouteTable struct {
	Routes Dll
}

func (dll *Dll) DllToRoute() *Layer3Route {
	return (*Layer3Route)(unsafe.Pointer(uintptr(unsafe.Pointer(dll)) - unsafe.Offsetof(Layer3Route{}.RouteGlue)))
}

func (header *IPHeader) Init() {
	header.Version = 4
	header.IHL = 5
	header.TOS = 0
	header.Length = 0
	header.ID = 0
	header.Flags = 0
	header.FragmentOffs = 0
	header.TTL = 64
	header.Protocol = 0
	header.Checksum = 0
	header.SourceIP = IPAddress{}
	header.DestinationIP = IPAddress{}
	header.Options = []byte{}
}

func (routingTable *Layer3RouteTable) LookupRoutingTableLPM(IP IPAddress) *Layer3Route {
	var route, lpmL3Route, defaultRoute *Layer3Route
	var subnet IPAddress
	var longestMask rune

	for dllRoute := routingTable.Routes.Next; dllRoute != nil; dllRoute = dllRoute.Next {
		route = dllRoute.DllToRoute()
		subnet = applyMask(IP, route.Mask)

		if bytes.Equal(constants.DefaultRoutingIp[:], route.DestinationIP[:]) && route.Mask == 0 {
			defaultRoute = route
		} else if bytes.Equal(subnet[:], route.DestinationIP[:]) {
			if route.Mask > longestMask {
				longestMask = route.Mask
				lpmL3Route = route
			}
		}
	}

	if lpmL3Route != nil {
		return lpmL3Route
	}
	return defaultRoute
}

func (routingTable *Layer3RouteTable) LookupRoutingTable(IP IPAddress, mask rune) *Layer3Route {
	for dllRoute := routingTable.Routes.Next; dllRoute != nil; dllRoute = dllRoute.Next {
		route := dllRoute.DllToRoute()
		if route.Mask == mask && bytes.Equal(IP[:], route.DestinationIP[:]) {
			return route
		}
	}
	return nil
}

func isRouteSame(route1 *Layer3Route, route2 *Layer3Route) bool {
	return route1.Mask == route2.Mask &&
		bytes.Equal(route1.DestinationIP[:], route2.DestinationIP[:]) &&
		bytes.Equal(route1.GatewayIP[:], route2.GatewayIP[:]) &&
		route1.IsDirect == route2.IsDirect &&
		bytes.Equal(route1.InterfaceName[:], route2.InterfaceName[:])
}

func (routingTable *Layer3RouteTable) DeleteRoute(IP IPAddress, mask rune) {
	route := routingTable.LookupRoutingTable(applyMask(IP, mask), mask)
	if route == nil {
		return
	}
	(&route.RouteGlue).RemoveNode()
}

func (routingTable *Layer3RouteTable) AddRoute(IP IPAddress, mask rune, gatewayIP *IPAddress, interfaceName *InterfaceName) {
	route := routingTable.LookupRoutingTableLPM(IP)

	if route != nil {
		fmt.Println("Error : Route already exists")
		return
	}
	route = &Layer3Route{
		DestinationIP: IPAddress{},
		Mask:          mask,
		IsDirect:      gatewayIP == nil && interfaceName == nil,
		GatewayIP:     IPAddress{},
		InterfaceName: InterfaceName{},
		RouteGlue:     Dll{},
	}

	subnet := applyMask(IP, mask)
	copy(route.DestinationIP[:], subnet[:])

	if gatewayIP != nil && interfaceName != nil {
		copy(route.GatewayIP[:], gatewayIP[:])
		copy(route.InterfaceName[:], interfaceName[:])
	}

	oldRoute := routingTable.LookupRoutingTable(route.DestinationIP, route.Mask)
	if oldRoute != nil && isRouteSame(oldRoute, route) {
		fmt.Println("Error : Route already exists")
		return
	}
	if oldRoute != nil {
		routingTable.DeleteRoute(oldRoute.DestinationIP, oldRoute.Mask)
	}
	(&route.RouteGlue).Init()
	(&routingTable.Routes).AddNode(&route.RouteGlue)
}

func (routingTable *Layer3RouteTable) AddDirectRoute(IP IPAddress, mask rune) {
	routingTable.AddRoute(IP, mask, nil, nil)
}

func (routingTable *Layer3RouteTable) Print() {
	for dllNode := routingTable.Routes.Next; dllNode != nil; dllNode = dllNode.Next {
		route := dllNode.DllToRoute()
		fmt.Println("Destination IP: ", route.DestinationIP.String(), ", Mask: ", route.Mask, ", IsDirect: ", route.IsDirect, ", Gateway IP: ", route.GatewayIP.String(), ", Interface Name: ", route.InterfaceName.String())
	}
}

func (header IPHeader) SerializeIPHeader() []byte {
	data := make([]byte, unsafe.Sizeof(header))
	data[0] = (header.Version << 4) | header.IHL
	data[1] = header.TOS
	binary.BigEndian.PutUint16(data[2:4], header.Length)
	binary.BigEndian.PutUint16(data[4:6], header.ID)
	data[6] = (header.Flags << 5) | uint8(header.FragmentOffs>>8)
	data[7] = uint8(header.FragmentOffs)
	data[8] = header.TTL
	data[9] = header.Protocol
	binary.BigEndian.PutUint16(data[10:12], header.Checksum)
	copy(data[12:28], header.SourceIP[:])
	copy(data[28:44], header.DestinationIP[:])
	copy(data[44:], header.Options)
	return data
}

func DeserializeIPHeader(data []byte) IPHeader {
	var header IPHeader
	header.Version = data[0] >> 4
	header.IHL = data[0] & 0x0F
	header.TOS = data[1]
	header.Length = binary.BigEndian.Uint16(data[2:4])
	header.ID = binary.BigEndian.Uint16(data[4:6])
	header.Flags = data[6] >> 5
	header.FragmentOffs = binary.BigEndian.Uint16(data[6:8]) & 0x1FFF
	header.TTL = data[8]
	header.Protocol = data[9]
	header.Checksum = binary.BigEndian.Uint16(data[10:12])
	copy(header.SourceIP[:], data[12:28])
	copy(header.DestinationIP[:], data[28:44])
	header.Options = append([]byte(nil), data[44:]...)
	return header
}
