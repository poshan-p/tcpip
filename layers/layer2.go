package layers

import (
	"bytes"
	"fmt"
	"tcpip/cmd/communication/send"
	"tcpip/constants"
	"tcpip/data"
)

func IsFrameReceivedOnIntfQualifying(intf *data.Interface, packet data.Packet, outputVLANID *uint) bool {
	*outputVLANID = 0

	var ethernetHeader *data.EthernetHeader

	vlanEthernetHeader := IsPacketVLANTagged(packet)
	if vlanEthernetHeader == nil {
		ethernetHeader = packet.DeserializeEthernetHeader()
	}

	if !intf.Properties.IsIpConfigured && intf.Properties.IntfL2Mode == constants.L2ModeUnknown {
		return false
	}

	if intf.Properties.IntfL2Mode == constants.ACCESS && intf.Properties.Vlans[0] == 0 {
		if vlanEthernetHeader == nil {
			return true
		} else {
			return false
		}
	}

	var intfVLANID, packetVlanID uint

	if intf.Properties.IntfL2Mode == constants.ACCESS {
		intfVLANID = intf.Properties.Vlans[0]

		if vlanEthernetHeader == nil && intf.Properties.Vlans[0] != 0 {
			*outputVLANID = intf.Properties.Vlans[0]
			return true
		}

		if vlanEthernetHeader == nil && intf.Properties.Vlans[0] == 0 {
			return true
		}

		packetVlanID = uint(vlanEthernetHeader.Tag.TCI & 0xFFF)

		if packetVlanID == intfVLANID {
			return true
		} else {
			return false
		}
	}

	if intf.Properties.IntfL2Mode == constants.TRUNK {
		if vlanEthernetHeader == nil {
			return false
		}
	}

	if intf.Properties.IntfL2Mode == constants.TRUNK && vlanEthernetHeader != nil {
		packetVlanID = uint(vlanEthernetHeader.Tag.TCI & 0xFFF)
		if intf.IsTrunkInterfaceVLANEnabled(packetVlanID) {
			return true
		} else {
			return false
		}
	}

	if intf.Properties.IsIpConfigured && vlanEthernetHeader != nil {
		return false
	}

	if intf.Properties.IsIpConfigured && bytes.Equal(intf.Properties.MAC[:], ethernetHeader.DestinationMAC[:]) {
		return true
	}

	if intf.Properties.IsIpConfigured && bytes.Equal(constants.BroadcastMacAddress[:], ethernetHeader.DestinationMAC[:]) {
		return true
	}
	return false
}

func SendARPBroadcastRequest(node *data.Node, oif *data.Interface, IP data.IPAddress) {
	if oif == nil {
		oif = node.GetMatchingSubnetInterface(IP)
		if oif == nil {
			fmt.Println("No eligible subnet for ARP resolution")
			return
		}
	}
	arpHeader := data.ArpHeader{
		HardwareType:          1,
		ProtocolType:          constants.EthernetIpProto,
		HardwareAddressLength: 6,
		ProtocolAddressLength: 4,
		OpCode:                constants.ArpBroadcastRequest,
	}
	copy(arpHeader.SourceMAC[:], oif.Properties.MAC[:])
	copy(arpHeader.SourceIP[:], oif.Properties.IP[:])
	copy(arpHeader.DestinationIP[:], IP[:])

	ethernetHeader := &data.EthernetHeader{
		Type: constants.ArpMessage,
	}

	copy(ethernetHeader.DestinationMAC[:], constants.BroadcastMacAddress[:])
	copy(ethernetHeader.SourceMAC[:], oif.Properties.MAC[:])
	copy(ethernetHeader.Payload[:], arpHeader.SerializeArpHeader())

	send.PacketSend((*ethernetHeader).SerializeEthernetHeader(), oif)
}

func sendARPReplyMessage(ethernetHeaderIn *data.EthernetHeader, oif *data.Interface) {
	arpHeaderIn := data.DeserializeArpHeader(ethernetHeaderIn.Payload[:])

	arpHeader := data.ArpHeader{
		HardwareType:          1,
		ProtocolType:          constants.EthernetIpProto,
		HardwareAddressLength: 6,
		ProtocolAddressLength: 4,
		OpCode:                constants.ArpReply,
	}
	copy(arpHeader.SourceIP[:], oif.Properties.IP[:])
	copy(arpHeader.DestinationIP[:], arpHeaderIn.SourceIP[:])
	copy(arpHeader.SourceMAC[:], oif.Properties.MAC[:])
	copy(arpHeader.DestinationMAC[:], arpHeaderIn.SourceMAC[:])

	ethernetHeader := &data.EthernetHeader{
		Type: constants.ArpMessage,
	}
	copy(ethernetHeader.DestinationMAC[:], arpHeaderIn.SourceMAC[:])
	copy(ethernetHeader.SourceMAC[:], oif.Properties.MAC[:])
	copy(ethernetHeader.Payload[:], arpHeader.SerializeArpHeader())

	send.PacketSend((*ethernetHeader).SerializeEthernetHeader(), oif)
}

func processARPReplyMessage(node *data.Node, iif *data.Interface, ethernetHdr *data.EthernetHeader) {
	fmt.Println("processARPReplyMessage: ARP reply message received on interface", iif.Name.String(),"of node", node.NodeName)

	UpdateFromArpReply(node.Properties.ArpTable, data.DeserializeArpHeader(ethernetHdr.Payload[:]), iif)
}

func processARPBroadcastRequest(node *data.Node, iif *data.Interface, ethernetHdr *data.EthernetHeader) {
	fmt.Println("processARPBroadcastRequest: ARP broadcast request message received on interface", iif.Name.String(), "of node", node.NodeName)

	arpHdr := data.DeserializeArpHeader(ethernetHdr.Payload[:])

	if !bytes.Equal(iif.Properties.IP[:], arpHdr.DestinationIP[:]) {
		fmt.Println("processARPBroadcastRequest: ARP Broadcast request message dropped, Destination IP address did not match")
		return
	}

	sendARPReplyMessage(ethernetHdr, iif)
}

func FrameReceive(node *data.Node, intf *data.Interface, packet data.Packet) {
	var vlanIdToTag uint = 0
	ethernetHdr := packet.DeserializeEthernetHeader()

	if !IsFrameReceivedOnIntfQualifying(intf, packet, &vlanIdToTag) {
		fmt.Println(node.NodeName, "rejected L2 Frame")
		return
	}

	fmt.Println(node.NodeName, " accepted L2 Frame")

	if intf.Properties.IsIpConfigured {
		switch ethernetHdr.Type {
		case constants.ArpMessage:
			arpHdr := data.DeserializeArpHeader(ethernetHdr.Payload[:])
			switch arpHdr.OpCode {
			case constants.ArpBroadcastRequest:
				processARPBroadcastRequest(node, intf, ethernetHdr)
			case constants.ArpReply:
				processARPReplyMessage(node, intf, ethernetHdr)
			default:
				break
			}
		default:
			PacketPromoteToLayer3(node, ethernetHdr.Payload, ethernetHdr.Type)
		}
	} else if intf.Properties.IntfL2Mode == constants.ACCESS || intf.Properties.IntfL2Mode == constants.TRUNK {
		if vlanIdToTag != 0 {
			taggedPacket := TagPacketWithVLANId(packet, vlanIdToTag).SerializeVLANEthernetHeader()
			copy(packet[:], taggedPacket[:])
		}
		SwitchFrameReceive(intf, packet)
	} else {
		return //drop packet
	}

}

func FrameReceiveFromTop(node *data.Node, gatewayIP data.IPAddress, intf *data.Interface, payload data.Payload, protocolNumber uint16) {
	if protocolNumber == constants.EthernetIpProto {
		ethernetHeader := &data.EthernetHeader{
			Type: constants.EthernetIpProto,
		}
		copy(ethernetHeader.Payload[:], payload[:])
		ForwardFrame(node, gatewayIP, intf, ethernetHeader)
	}
}

func ForwardFrame(node *data.Node, gatewayIP data.IPAddress, intf *data.Interface, ethernetHeader *data.EthernetHeader) {
	var oif *data.Interface
	var arpEntry *data.ArpEntry

	if intf != nil {
		oif = intf

		arpEntry = data.ArpTableLookup(node.Properties.ArpTable, gatewayIP)
		if arpEntry == nil {
			CreateArpSaneEntry(node.Properties.ArpTable, gatewayIP, ethernetHeader.SerializeEthernetHeader())
			SendARPBroadcastRequest(node, oif, gatewayIP)
			return
		} else {
			goto framePrepare
		}
	}
	if IsRouteLocalDelivery(node, gatewayIP) {
		PacketPromoteToLayer3(node, ethernetHeader.Payload, ethernetHeader.Type)
		return
	}

	oif = node.GetMatchingSubnetInterface(gatewayIP)
	if oif == nil {
		fmt.Println("No eligible subnet for ARP resolution")
		return
	}

	if bytes.Equal(oif.Properties.IP[:], gatewayIP[:]) {
		copy(ethernetHeader.SourceMAC[:], oif.Properties.MAC[:])
		copy(ethernetHeader.DestinationMAC[:], oif.Properties.MAC[:])
		send.PacketSendToSelf(ethernetHeader.SerializeEthernetHeader(), oif)
		return
	}

	arpEntry = data.ArpTableLookup(node.Properties.ArpTable, gatewayIP)

	if arpEntry == nil || (arpEntry != nil && arpEntry.IsSane) {
		CreateArpSaneEntry(node.Properties.ArpTable, gatewayIP, ethernetHeader.SerializeEthernetHeader())
		SendARPBroadcastRequest(node, oif, gatewayIP)
		return
	}

framePrepare:
	copy(ethernetHeader.SourceMAC[:], oif.Properties.MAC[:])
	copy(ethernetHeader.DestinationMAC[:], arpEntry.MAC[:])
	send.PacketSend(ethernetHeader.SerializeEthernetHeader(), oif)
}

func pendingArpEntryCallback(node *data.Node, intf *data.Interface, arpEntry *data.ArpEntry, arpPendingEntry *data.ArpPendingEntry) {
	ethernetHeader := arpPendingEntry.Packet.DeserializeEthernetHeader()
	copy(ethernetHeader.DestinationMAC[:], arpEntry.MAC[:])
	copy(ethernetHeader.SourceMAC[:], intf.Properties.MAC[:])
	send.PacketSend(ethernetHeader.SerializeEthernetHeader(), intf)
}

func CreateArpSaneEntry(arpTable *data.ArpTable, ipAddress data.IPAddress, packet data.Packet) {
	arpEntry := data.ArpTableLookup(arpTable, ipAddress)

	if arpEntry != nil {
		if !arpEntry.IsSane {
			panic("Arp Entry is not sane")
		}
		arpEntry.AddPendingArpTableEntry(pendingArpEntryCallback, packet)
		return
	}
	arpEntry = &data.ArpEntry{}
	arpEntry.ArpGlue.Init()
	copy(arpEntry.IP[:], ipAddress[:])
	arpEntry.IsSane = true
	(*arpEntry).AddPendingArpTableEntry(pendingArpEntryCallback, packet)
	if !data.AddArpTableEntry(arpTable, arpEntry, nil) {
		panic("Arp Entry already exists")
	}
}

func UpdateFromArpReply(arpTable *data.ArpTable, arpHeader *data.ArpHeader, intf *data.Interface) {
	if arpHeader.OpCode != constants.ArpReply {
		panic("Not an Arp Reply")
	}
	var arpPendingList *data.Dll = nil
	arpEntry := &data.ArpEntry{
		IsSane: false,
	}

	copy(arpEntry.IP[:], arpHeader.SourceIP[:])
	copy(arpEntry.MAC[:], arpHeader.SourceMAC[:])
	copy(arpEntry.InterfaceName[:], intf.Name[:])

	rc := data.AddArpTableEntry(arpTable, arpEntry, &arpPendingList)
	if arpPendingList != nil {
		for dllArpPendingEntry := arpPendingList.Next; dllArpPendingEntry != nil; dllArpPendingEntry = dllArpPendingEntry.Next {
			arpPendingEntry := dllArpPendingEntry.DllToArpPendingEntry()
			arpPendingEntry.ArpCallback(intf.Node, intf, arpEntry, arpPendingEntry)
			arpPendingEntry.ArpPendingEntryGlue.RemoveNode()
		}
		arpPendingList.DllToArpEntry().IsSane = false
	}
	if !rc {
		arpEntry.DeleteArpEntry()
	}
}

func SetIntfL2Mode(node *data.Node, intfName string, mode int) {

	intf := node.GetNodeIntfByName(intfName)

	if mode == constants.L2ModeUnknown {
		panic("Invalid L2 mode option")
	}

	if intf.Properties.IsIpConfigured {
		intf.Properties.IsIpConfigured = false
		intf.Properties.IsIpConfigBackup = true
		intf.Properties.IntfL2Mode = mode
		return
	}

	if intf.Properties.IntfL2Mode == constants.L2ModeUnknown {
		intf.Properties.IntfL2Mode = mode
		return
	}

	if intf.Properties.IntfL2Mode == mode {
		return
	}

	if intf.Properties.IntfL2Mode == constants.ACCESS && mode == constants.TRUNK {
		intf.Properties.IntfL2Mode = mode
		return
	}

	if intf.Properties.IntfL2Mode == constants.TRUNK && mode == constants.ACCESS {
		intf.Properties.IntfL2Mode = mode

		for i := 0; i < int(constants.MaxVlanMembership); i++ {
			intf.Properties.Vlans[i] = 0
		}
	}
}

func SetIntfVLAN(node *data.Node, intfName string, vlanID uint) {
	intf := node.GetNodeIntfByName(intfName)

	if intf.Properties.IsIpConfigured {
		fmt.Printf("Error: Interface %s: L3 mode enabled\n", intf.Name.String())
		return
	}

	if intf.Properties.IntfL2Mode != constants.ACCESS && intf.Properties.IntfL2Mode != constants.TRUNK {
		fmt.Printf("Error: Interface %s: L2 mode not enabled\n", intf.Name.String())
		return
	}

	if intf.Properties.IntfL2Mode == constants.ACCESS {
		for i, v := range intf.Properties.Vlans {
			if v != 0 {
				intf.Properties.Vlans[i] = vlanID
				return
			}
		}
		intf.Properties.Vlans[0] = vlanID
		return
	}

	if intf.Properties.IntfL2Mode == constants.TRUNK {
		for i, v := range intf.Properties.Vlans {
			if v == 0 {
				intf.Properties.Vlans[i] = vlanID
				return
			} else if v == vlanID {
				return
			}
		}
		fmt.Printf("Error: Interface %s: Max VLAN membership limit reached\n", intf.Name.String())
	}
}

func IsPacketVLANTagged(packet data.Packet) *data.VLANEthernetHeader {
	vlanEthernetHeader := packet.DeserializeVLANEthernetHeader()

	if vlanEthernetHeader.Tag.TPID == constants.Vlan8021qProto {
		return vlanEthernetHeader
	}
	return nil
}

func TagPacketWithVLANId(packet data.Packet, vlanID uint) *data.VLANEthernetHeader {
	vlanEthernetHeader := IsPacketVLANTagged(packet)

	if vlanEthernetHeader != nil {
		vlanEthernetHeader.Tag = data.SetVLAN(vlanID)
		return vlanEthernetHeader
	}

	ethernetHeader := packet.DeserializeEthernetHeader()
	vlanEthernetHeader = &data.VLANEthernetHeader{
		Tag:  data.SetVLAN(vlanID),
		Type: ethernetHeader.Type,
		FCS:  ethernetHeader.FCS,
	}

	copy(vlanEthernetHeader.DestinationMAC[:], ethernetHeader.DestinationMAC[:])
	copy(vlanEthernetHeader.SourceMAC[:], ethernetHeader.SourceMAC[:])
	copy(vlanEthernetHeader.Payload[:], ethernetHeader.Payload[:])

	return vlanEthernetHeader
}

func UntagPacketWithVLANId(packet data.Packet) *data.EthernetHeader {
	vlanEthernetHeader := IsPacketVLANTagged(packet)
	if vlanEthernetHeader == nil {
		return packet.DeserializeEthernetHeader()
	}

	ethernetHeader := &data.EthernetHeader{
		Type: vlanEthernetHeader.Type,
		FCS:  vlanEthernetHeader.FCS,
	}

	copy(ethernetHeader.DestinationMAC[:], vlanEthernetHeader.DestinationMAC[:])
	copy(ethernetHeader.SourceMAC[:], vlanEthernetHeader.SourceMAC[:])
	copy(ethernetHeader.Payload[:], vlanEthernetHeader.Payload[:])

	return ethernetHeader
}

func MacTableLookup(macTable *data.MacTable, mac data.MacAddress) *data.MacEntry {
	for dllMacEntry := macTable.MacEntries.Next; dllMacEntry != nil; dllMacEntry = dllMacEntry.Next {
		macEntry := dllMacEntry.DllToMacEntry()
		if bytes.Equal(macEntry.MAC[:], mac[:]) {
			return macEntry
		}
	}
	return nil
}

func DeleteMacTableEntry(macTable *data.MacTable, mac data.MacAddress) {
	macEntry := MacTableLookup(macTable, mac)
	if macEntry == nil {
		return
	}
	(&macEntry.MacGlue).RemoveNode()
}

func AddMacTableEntry(macTable *data.MacTable, macEntry *data.MacEntry) bool {
	oldEntry := MacTableLookup(macTable, macEntry.MAC)

	if oldEntry != nil && bytes.Equal(oldEntry.MAC[:], macEntry.MAC[:]) && bytes.Equal(oldEntry.InterfaceName[:], macEntry.InterfaceName[:]) {
		return false
	}

	if oldEntry != nil {
		DeleteMacTableEntry(macTable, oldEntry.MAC)
	}

	(&macEntry.MacGlue).Init()
	(&macTable.MacEntries).AddNode(&macEntry.MacGlue)
	return true
}

func SwitchFrameReceive(intf *data.Interface, packet data.Packet) {
	node := intf.Node
	ethernetHeader := packet.DeserializeEthernetHeader()
	fmt.Print("src: ", ethernetHeader.SourceMAC.String(), " dest: ", ethernetHeader.DestinationMAC.String(), "\n")

	switchMacLearning(node, ethernetHeader.SourceMAC, intf.Name)
	switchFrameForward(node, intf, packet)
}

func switchMacLearning(node *data.Node, sourceMAC data.MacAddress, intfName [16]byte) {
	macEntry := &data.MacEntry{}

	copy(macEntry.InterfaceName[:], intfName[:])
	copy(macEntry.MAC[:], sourceMAC[:])
	AddMacTableEntry(node.Properties.MacTable, macEntry)
}

func switchFrameForward(node *data.Node, intf *data.Interface, packet data.Packet) {
	ethernetHeader := packet.DeserializeEthernetHeader()
	if bytes.Equal(ethernetHeader.DestinationMAC[:], constants.BroadcastMacAddress[:]) {
		FloodPacket(node, intf, packet)
		return
	}
	macEntry := MacTableLookup(node.Properties.MacTable, ethernetHeader.DestinationMAC)

	if macEntry == nil {
		FloodPacket(node, intf, packet)
		return
	}
	oif := node.GetNodeIntfByName(macEntry.InterfaceName.String())
	if oif == nil {
		return
	}
	SwitchSendPacketOut(packet, oif)
}

func FloodPacket(node *data.Node, excludedIntf *data.Interface, packet data.Packet) {
	for _, intf := range node.Interfaces {
		if intf == nil {
			return
		}
		if bytes.Equal(intf.Name[:], excludedIntf.Name[:]) {
			continue
		}
		SwitchSendPacketOut(packet, intf)
	}
}

func SwitchSendPacketOut(packet data.Packet, intf *data.Interface) bool {
	if intf.Properties.IsIpConfigured {
		panic("Invalid operation: Attempting to send a packet out of an L3 mode interface")
	}

	if intf.Properties.IntfL2Mode == constants.L2ModeUnknown {
		return false
	}

	vlanEthernetHeader := IsPacketVLANTagged(packet)

	switch intf.Properties.IntfL2Mode {
	case constants.ACCESS:

		if intf.Properties.Vlans[0] == 0 && vlanEthernetHeader == nil {
			send.PacketSend(packet, intf)
			return true
		}

		if intf.Properties.Vlans[0] != 0 && vlanEthernetHeader == nil {
			return false
		}

		if vlanEthernetHeader != nil && intf.Properties.Vlans[0] == uint(vlanEthernetHeader.Tag.TCI&0xFFF) {
			ethernetHeader := &data.EthernetHeader{
				Type: vlanEthernetHeader.Type,
				FCS:  vlanEthernetHeader.FCS,
			}
			copy(ethernetHeader.DestinationMAC[:], vlanEthernetHeader.DestinationMAC[:])
			copy(ethernetHeader.SourceMAC[:], vlanEthernetHeader.SourceMAC[:])
			copy(ethernetHeader.Payload[:], vlanEthernetHeader.Payload[:])

			send.PacketSend((*ethernetHeader).SerializeEthernetHeader(), intf)
			return true
		}

		if intf.Properties.Vlans[0] == 0 && vlanEthernetHeader != nil {
			return false
		}

	case constants.TRUNK:
		packetVlanID := uint(0)

		if vlanEthernetHeader != nil {
			packetVlanID = uint(vlanEthernetHeader.Tag.TCI & 0xFFF)
		}

		if packetVlanID != 0 && intf.IsTrunkInterfaceVLANEnabled(packetVlanID) {
			send.PacketSend(packet, intf)
			return true
		}

		return false

	case constants.L2ModeUnknown:
		return false

	default:
		panic("Invalid L2 mode")
	}
	return false
}
