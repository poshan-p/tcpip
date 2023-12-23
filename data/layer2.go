package data

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"tcpip/constants"
	"unsafe"
)

type EthernetHeader struct {
	DestinationMAC MacAddress
	SourceMAC      MacAddress
	Type           uint16
	Payload        Payload
	FCS            uint32
}

type VLANEthernetHeader struct {
	DestinationMAC MacAddress
	SourceMAC      MacAddress
	Tag            VLANTag
	Type           uint16
	Payload        Payload
	FCS            uint32
}

type VLANTag struct {
	TPID uint16
	TCI  uint16
}

type ArpHeader struct {
	HardwareType          uint16
	ProtocolType          uint16
	HardwareAddressLength uint8
	ProtocolAddressLength uint8
	OpCode                uint16
	SourceMAC             MacAddress
	SourceIP              IPAddress
	DestinationMAC        MacAddress
	DestinationIP         IPAddress
}

type ArpTable struct {
	ArpEntries Dll
}

type ArpEntry struct {
	IP             IPAddress
	MAC            MacAddress
	InterfaceName  InterfaceName
	IsSane         bool
	ArpPendingList Dll
	ArpGlue        Dll
}

type ArpPendingEntry struct {
	ArpCallback         ArpCallback
	Packet              Packet
	ArpPendingEntryGlue Dll
}

type ArpCallback func(*Node, *Interface, *ArpEntry, *ArpPendingEntry)

type MacTable struct {
	MacEntries Dll
}

type MacEntry struct {
	MAC           MacAddress
	InterfaceName InterfaceName
	MacGlue       Dll
}

func (dll *Dll) DllToArpEntry() *ArpEntry {
	return (*ArpEntry)(unsafe.Pointer(uintptr(unsafe.Pointer(dll)) - unsafe.Offsetof(ArpEntry{}.ArpGlue)))
}

func (dll *Dll) DllToArpPendingEntry() *ArpPendingEntry {
	return (*ArpPendingEntry)(unsafe.Pointer(uintptr(unsafe.Pointer(dll)) - unsafe.Offsetof(ArpPendingEntry{}.ArpPendingEntryGlue)))
}

func (dll *Dll) DllToMacEntry() *MacEntry {
	return (*MacEntry)(unsafe.Pointer(uintptr(unsafe.Pointer(dll)) - unsafe.Offsetof(MacEntry{}.MacGlue)))
}

func SetVLAN(vlanID uint) VLANTag {
	if vlanID < 0 || vlanID > 4095 {
		panic("Invalid VLAN ID. It must be in the range 0-4095.")
	}
	// Set the TPID (Tag Protocol Identifier)
	tpid := constants.Vlan8021qProto // 0x8100 represents the IEEE 802.1Q protocol
	// Set the TCI (Tag Control Information)
	tci := uint16(vlanID & 0xFFF) // Use the lowest 12 bits for VLAN ID
	// Combine TPID and TCI to form the 802.1Q Tag
	tag := (tpid << 12) | tci
	// Create the VLAN8021QHeader structure
	vlanHeader := VLANTag{
		TPID: tpid,
		TCI:  tag,
	}
	return vlanHeader
}

func (hdr *VLANTag) GetVlanID() uint16 {
	// Extract VLAN ID (bits 0-11 of TCI)
	return hdr.TCI & 0xFFF
}

func (packet PacketWithAux) ExtractAuxAndData() (InterfaceName, Packet) {
	var intfName [constants.MaxAuxiliarySize]byte
	var packetData Packet
	copy(intfName[:], packet[:constants.MaxAuxiliarySize])
	copy(packetData[:], packet[constants.MaxAuxiliarySize:])
	return intfName, packetData
}

func (header EthernetHeader) SerializeEthernetHeader() Packet {
	data := Packet{}
	copy(data[0:6], header.DestinationMAC[:])
	copy(data[6:12], header.SourceMAC[:])
	binary.BigEndian.PutUint16(data[12:14], header.Type)
	copy(data[14:], header.Payload[:])
	binary.BigEndian.PutUint32(data[len(data)-4:], header.FCS)
	return data
}

func (data Packet) DeserializeEthernetHeader() *EthernetHeader {
	var header EthernetHeader
	copy(header.DestinationMAC[:], data[0:6])
	copy(header.SourceMAC[:], data[6:12])
	header.Type = binary.BigEndian.Uint16(data[12:14])
	copy(header.Payload[:], data[14:len(data)-4])
	header.FCS = binary.BigEndian.Uint32(data[len(data)-4:])
	return &header
}

func (header ArpHeader) SerializeArpHeader() ArpHeaderBytes {
	data := make([]byte, binary.Size(header))
	binary.BigEndian.PutUint16(data[0:2], header.HardwareType)
	binary.BigEndian.PutUint16(data[2:4], header.ProtocolType)
	data[4] = header.HardwareAddressLength
	data[5] = header.ProtocolAddressLength
	binary.BigEndian.PutUint16(data[6:8], header.OpCode)
	copy(data[8:14], header.SourceMAC[:])
	copy(data[14:30], header.SourceIP[:])
	copy(data[30:36], header.DestinationMAC[:])
	copy(data[36:52], header.DestinationIP[:])
	return data
}

func DeserializeArpHeader(data ArpHeaderBytes) *ArpHeader {
	var header ArpHeader
	header.HardwareType = binary.BigEndian.Uint16(data[0:2])
	header.ProtocolType = binary.BigEndian.Uint16(data[2:4])
	header.HardwareAddressLength = data[4]
	header.ProtocolAddressLength = data[5]
	header.OpCode = binary.BigEndian.Uint16(data[6:8])
	copy(header.SourceMAC[:], data[8:14])
	copy(header.SourceIP[:], data[14:30])
	copy(header.DestinationMAC[:], data[30:36])
	copy(header.DestinationIP[:], data[36:52])
	return &header
}

func (header VLANEthernetHeader) SerializeVLANEthernetHeader() Packet {
	data := Packet{}
	copy(data[0:6], header.DestinationMAC[:])
	copy(data[6:12], header.SourceMAC[:])
	binary.BigEndian.PutUint16(data[12:14], header.Tag.TPID)
	binary.BigEndian.PutUint16(data[14:16], header.Tag.TCI)
	binary.BigEndian.PutUint16(data[16:18], uint16(header.Type))
	copy(data[18:], header.Payload[:])
	binary.BigEndian.PutUint32(data[len(data)-4:], header.FCS)
	return data
}

func (data Packet) DeserializeVLANEthernetHeader() *VLANEthernetHeader {
	var header VLANEthernetHeader
	copy(header.DestinationMAC[:], data[0:6])
	copy(header.SourceMAC[:], data[6:12])
	header.Tag.TPID = binary.BigEndian.Uint16(data[12:14])
	header.Tag.TCI = binary.BigEndian.Uint16(data[14:16])
	header.Type = binary.BigEndian.Uint16(data[16:18])
	copy(header.Payload[:], data[18:len(data)-4])
	header.FCS = binary.BigEndian.Uint32(data[len(data)-4:])
	return &header
}

func ArpTableLookup(arpTable *ArpTable, IP IPAddress) *ArpEntry {
	for dllArpEntry := arpTable.ArpEntries.Next; dllArpEntry != nil; dllArpEntry = dllArpEntry.Next {
		arpEntry := dllArpEntry.DllToArpEntry()
		if bytes.Equal(arpEntry.IP[:], IP[:]) {
			return arpEntry
		}
	}
	return nil
}

func AddArpTableEntry(arpTable *ArpTable, arpEntry *ArpEntry, arpPendingList **Dll) bool {
	if arpPendingList != nil && *arpPendingList != nil {
		panic("arpPendingList is not nil")
	}

	oldEntry := ArpTableLookup(arpTable, arpEntry.IP)

	if oldEntry == nil {
		arpTable.ArpEntries.AddNode(&arpEntry.ArpGlue)
		return true
	}
	if oldEntry != nil && IsArpEntriesEqual(oldEntry, arpEntry) {
		return false
	}

	if oldEntry != nil && !oldEntry.IsSane {
		oldEntry.DeleteArpEntry()
		arpEntry.ArpGlue.Init()
		arpTable.ArpEntries.AddNode(&arpEntry.ArpGlue)
		return true
	}

	if oldEntry != nil && oldEntry.IsSane && arpEntry.IsSane {
		if !arpEntry.ArpPendingList.IsDLLEmpty() {
			oldEntry.ArpPendingList.AddNode(arpEntry.ArpPendingList.Next)
		}

		if arpPendingList != nil {
			*arpPendingList = &oldEntry.ArpPendingList
		}
		return false
	}

	if oldEntry != nil && oldEntry.IsSane && !arpEntry.IsSane {
		copy(oldEntry.MAC[:], arpEntry.MAC[:])
		copy(oldEntry.InterfaceName[:], arpEntry.InterfaceName[:])

		if arpPendingList != nil {
			*arpPendingList = &oldEntry.ArpPendingList
		}
		return false
	}
	return false
}

func (arpEntry *ArpEntry) DeleteArpEntry() {
	(&arpEntry.ArpGlue).RemoveNode()
	for dllArpPendingEntry := arpEntry.ArpPendingList.Next; dllArpPendingEntry != nil; dllArpPendingEntry = dllArpPendingEntry.Next {
		arpPendingEntry := dllArpPendingEntry.DllToArpPendingEntry()
		(&arpPendingEntry.ArpPendingEntryGlue).RemoveNode()
	}
}

func IsArpEntriesEqual(arpEntry1 *ArpEntry, arpEntry2 *ArpEntry) bool {
	if arpEntry1 == nil || arpEntry2 == nil {
		return false
	}
	return bytes.Equal(arpEntry1.IP[:], arpEntry2.IP[:]) &&
		bytes.Equal(arpEntry1.MAC[:], arpEntry2.MAC[:]) &&
		bytes.Equal(arpEntry1.InterfaceName[:], arpEntry2.InterfaceName[:]) &&
		arpEntry1.IsSane == arpEntry2.IsSane
}

func (arpEntry *ArpEntry) AddPendingArpTableEntry(callback ArpCallback, packet Packet) {
	arpPendingEntry := &ArpPendingEntry{}
	arpPendingEntry.ArpPendingEntryGlue.Init()
	arpPendingEntry.ArpCallback = callback
	copy(arpPendingEntry.Packet[:], packet[:])
	(&arpEntry.ArpPendingList).AddNode(&arpPendingEntry.ArpPendingEntryGlue)

}

func (arpTable *ArpTable) Print() {
	for dllArpEntry := arpTable.ArpEntries.Next; dllArpEntry != nil; dllArpEntry = dllArpEntry.Next {
		arpEntry := dllArpEntry.DllToArpEntry()

		fmt.Printf("IP: %s, Mac: %s, Interface: %v\n", arpEntry.IP.String(), arpEntry.MAC.String(), arpEntry.InterfaceName.String())
	}
}

func (macTable *MacTable) Print() {
	for dllMacEntry := macTable.MacEntries.Next; dllMacEntry != nil; dllMacEntry = dllMacEntry.Next {
		macEntry := dllMacEntry.DllToMacEntry()

		fmt.Printf("Mac: %s, Interface: %v\n", macEntry.MAC.String(), macEntry.InterfaceName.String())
	}
}

func (packet Packet) IsPacketVLANTagged() *VLANEthernetHeader {
	vlanEthernetHeader := packet.DeserializeVLANEthernetHeader()

	if vlanEthernetHeader.Tag.TPID == constants.Vlan8021qProto {
		return vlanEthernetHeader
	}
	return nil
}
