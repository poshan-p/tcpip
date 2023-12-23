package layers

import (
	"bytes"
	"fmt"
	"tcpip/constants"
	"tcpip/data"
	"unsafe"
)

func PacketReceive(node *data.Node, payload data.Payload) {
	ipHeader := data.DeserializeIPHeader(payload[:unsafe.Sizeof(data.IPHeader{})])

	route := node.Properties.RoutingTable.LookupRoutingTableLPM(ipHeader.DestinationIP)
	if route == nil {
		fmt.Println("No route found")
		return
	}
	if route.IsDirect {
		if IsRouteLocalDelivery(node, ipHeader.DestinationIP) {
			switch ipHeader.Protocol {
			case constants.IcmpProto:
				fmt.Println("IP Address: ", ipHeader.DestinationIP.String(), ", ping received")
			case constants.IpInIpProto:
				_payload := data.Payload{}
				copy(_payload[:], payload[unsafe.Sizeof(data.IPHeader{}):])
				PacketReceive(node, _payload)
			default:
				break
			}

		} else {
			PacketDemoteToLayer2(node, data.IPAddress{}, nil, payload, constants.EthernetIpProto)
		}
	} else {
		ipHeader.TTL--
		if ipHeader.TTL == 0 {
			fmt.Println("TTL expired")
			return
		}
		PacketDemoteToLayer2(node, route.GatewayIP, node.GetNodeIntfByName(route.InterfaceName.String()), payload, constants.EthernetIpProto)
	}
}

func PacketReceiveFromTop(node *data.Node, appData []byte, protocolNumber uint8, destinationIP data.IPAddress) {
	ipHeader := &data.IPHeader{}
	ipHeader.Init()

	ipHeader.Protocol = protocolNumber
	copy(ipHeader.DestinationIP[:], destinationIP[:])
	copy(ipHeader.SourceIP[:], node.Properties.LB[:])

	ipHeader.IHL = uint8(unsafe.Sizeof(data.IPHeader{}) / 4)

	route := node.Properties.RoutingTable.LookupRoutingTableLPM(ipHeader.DestinationIP)
	if route == nil {
		fmt.Println("No route found")
		return
	}
	var payload data.Payload
	gatewayIP := data.IPAddress{}

	if !route.IsDirect {
		copy(gatewayIP[:], route.GatewayIP[:])
	} else {
		copy(gatewayIP[:], ipHeader.DestinationIP[:])
	}

	ipHeaderSerialized := ipHeader.SerializeIPHeader()
	copy(payload[:], ipHeaderSerialized[:])
	copy(payload[unsafe.Sizeof(data.IPHeader{}):], appData[:])

	if route.IsDirect {
		fmt.Println("Route is direct")
		PacketDemoteToLayer2(node, gatewayIP, nil, payload, constants.EthernetIpProto)
	} else {
		PacketDemoteToLayer2(node, gatewayIP, node.GetNodeIntfByName(route.InterfaceName.String()), payload, constants.EthernetIpProto)
	}

}

func IsRouteLocalDelivery(node *data.Node, destinationIP data.IPAddress) bool {
	if node.Properties.IsLbConfigured && bytes.Equal(destinationIP[:], node.Properties.LB[:]) {
		return true
	}

	for _, intf := range node.Interfaces {
		if intf == nil {
			return false
		}
		if !intf.Properties.IsIpConfigured {
			continue
		}
		if bytes.Equal(destinationIP[:], intf.Properties.IP[:]) {
			return true

		}
	}
	return false
}
