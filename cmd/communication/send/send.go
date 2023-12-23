package send

import (
	"fmt"
	"net"
	"tcpip/constants"
	"tcpip/data"
)

// PacketSend sends a packet out through the specified interface to the connected neighbor node.
func PacketSend(packet data.Packet, intf *data.Interface) int {
	var intfName data.InterfaceName
	neighbourNode := intf.GetNeighbourNode()

	if intf.Link.Interface1 == *intf {
		intfName = intf.Link.Interface2.Name
	} else {
		intfName = intf.Link.Interface1.Name
	}

	// Resolve the destination UDP address
	destAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", "127.0.0.1", neighbourNode.UDPPortNumber))
	if err != nil {
		fmt.Printf("Error resolving destination UDP address: %v", err)
		return -1
	}

	// Create a UDP connection each time a packet needs to be sent
	conn, err := net.DialUDP("udp", nil, destAddr)
	if err != nil {
		fmt.Printf("Error creating UDP connection: %v", err)
		return -1
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}(conn)

	var packetWithAux data.PacketWithAux
	copy(packetWithAux[:constants.MaxAuxiliarySize], intfName[:])
	copy(packetWithAux[constants.MaxAuxiliarySize:], packet[:])

	// Send the packet to the destination
	_, err = conn.Write(packetWithAux[:])
	if err != nil {
		fmt.Printf("Error sending packet: %v", err)
		return -1
	}

	return 0
}

func PacketSendToSelf(packet data.Packet, intf *data.Interface) {
	neighbourNode := intf.Node

	// Resolve the destination UDP address
	destAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", "127.0.0.1", neighbourNode.UDPPortNumber))
	if err != nil {
		fmt.Printf("Error resolving destination UDP address: %v", err)
		return
	}

	// Create a UDP connection each time a packet needs to be sent
	conn, err := net.DialUDP("udp", nil, destAddr)
	if err != nil {
		fmt.Printf("Error creating UDP connection: %v", err)
		return
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}(conn)

	var packetWithAux data.PacketWithAux
	copy(packetWithAux[:constants.MaxAuxiliarySize], intf.Name[:])
	copy(packetWithAux[constants.MaxAuxiliarySize:], packet[:])

	// Send the packet to the destination
	_, err = conn.Write(packetWithAux[:])
	if err != nil {
		fmt.Printf("Error sending packet: %v", err)
		return
	}
}
