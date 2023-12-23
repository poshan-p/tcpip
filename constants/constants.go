package constants

const (
	MaxIntfPerNode      int  = 10
	MaxPacketBufferSize int  = 1540 // 1536
	MaxPacketSize       int  = 1524 //1520
	MaxPayloadSize      int  = 1500
	MaxAuxiliarySize    int  = 16
	MaxVlanMembership   uint = 10
)

const (
	Vlan8021qProto  uint16 = 0x8100
	EthernetIpProto uint16 = 0x0800
	IcmpProto       uint8  = 0x01
	IpInIpProto     uint8  = 0x04
)

const (
	ArpBroadcastRequest uint16 = 1
	ArpReply            uint16 = 2
	ArpMessage          uint16 = 806
)

var BroadcastMacAddress = [6]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

var DefaultRoutingIp = [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

const (
	ACCESS = iota
	TRUNK
	L2ModeUnknown
)
