package receive

import (
	"tcpip/data"
	"tcpip/layers"
)

func PacketReceive(node *data.Node, packetWithAux data.PacketWithAux) {
	intfName, packet := packetWithAux.ExtractAuxAndData()

	intf := node.GetNodeIntfByName(intfName.String())
	if intf == nil {
		panic("Interface not found")
	}

	layers.FrameReceive(node, intf, packet)
}
