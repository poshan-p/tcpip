package layers

import (
	"tcpip/constants"
	"tcpip/data"
)

func PacketDemoteToLayer3(node *data.Node, appData []byte, protocolNumber uint8, destinationIP data.IPAddress) {
	PacketReceiveFromTop(node, appData, protocolNumber, destinationIP)
}

func PacketDemoteToLayer2(node *data.Node, gatewayIP data.IPAddress, intf *data.Interface, payload data.Payload, protocolNumber uint16) {
	FrameReceiveFromTop(node, gatewayIP, intf, payload, protocolNumber)
}

func PacketPromoteToLayer3(node *data.Node, payload data.Payload, protocolNumber uint16) {
	switch protocolNumber {
	case constants.EthernetIpProto:
		PacketReceive(node, payload)
	default:
		break
	}
}
