package cmd

import (
	"bytes"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"strconv"
	"tcpip/constants"
	"tcpip/data"
	"tcpip/layers"
	"tcpip/topology"
	"unsafe"
)

var Topology = topology.SquareTopology()

func ShowTopology(c *cli.Context) {
	Topology.Print()
}

func ShowNodeArpTable(c *cli.Context) {
	nodeName := c.Args().First()
	if nodeName == "" {
		fmt.Println("Please provide a node name")
		return
	}
	node := (*Topology).GetNodeByName(nodeName)
	if node == nil {
		fmt.Println("Node not found")
		return
	}
	node.Properties.ArpTable.Print()
}

func ShowNodeMacTable(c *cli.Context) {
	nodeName := c.Args().First()
	if nodeName == "" {
		fmt.Println("Please provide a node name")
		return
	}
	node := (*Topology).GetNodeByName(nodeName)
	if node == nil {
		fmt.Println("Node not found")
		return
	}
	node.Properties.MacTable.Print()
}

func ShowNodeRoutingTable(c *cli.Context) {
	nodeName := c.Args().First()
	if nodeName == "" {
		fmt.Println("Please provide a node name")
		return
	}
	node := (*Topology).GetNodeByName(nodeName)
	if node == nil {
		fmt.Println("Node not found")
		return
	}
	node.Properties.RoutingTable.Print()
}

func ShowNode(c *cli.Context) {
	nodeName := c.Args().First()
	if nodeName == "" {
		fmt.Println("Please provide a node name")
		return
	}
	node := (*Topology).GetNodeByName(nodeName)
	if node == nil {
		fmt.Println("Node not found")
		return
	}
	node.Print()
}

func ExitCommand(c *cli.Context) {
	os.Exit(0)
}

func ShowHelp(c *cli.Context) {
	err := cli.ShowAppHelp(c)
	if err != nil {
		return
	}
}

func HandleCommandNotFound(c *cli.Context, command string) {
	fmt.Printf("Unrecognized command: %s\n", command)
	fmt.Println("Type 'help' to see a list of available commands.")
}

func RunResolveARPCommand(c *cli.Context) {
	emptyIPAddress := data.IPAddress{}
	nodeName := c.Args().Get(0)
	if nodeName == "" {
		fmt.Println("Invalid command structure. Use 'run node <nodeName> resolve-arp [command options] [arguments...]'")
		return
	}
	node := (*Topology).GetNodeByName(nodeName)
	if node == nil {
		fmt.Println("Node not found")
		return
	}
	ip := data.StringToIPAddress(c.Args().Get(1))
	if bytes.Equal(ip[:], emptyIPAddress[:]) {
		fmt.Println("Invalid IP address")
		return
	}
	layers.SendARPBroadcastRequest(node, nil, ip)
}

func RunPingCommand(c *cli.Context) {
	_nodeName := c.Args().Get(0)
	_gatewayIP := c.Args().Get(1)

	if _nodeName == "" || _gatewayIP == "" {
		fmt.Println("Invalid command structure. Use 'run node ping <nodeName> <gatewayIP>'")
		return
	}

	node := (*Topology).GetNodeByName(_nodeName)
	if node == nil {
		fmt.Println("Node not found")
		return
	}
	emptyIPAddress := data.IPAddress{}
	ip := data.StringToIPAddress(_gatewayIP)
	if bytes.Equal(ip[:], emptyIPAddress[:]) {
		fmt.Println("Invalid IP address")
		return
	}

	layers.PacketDemoteToLayer3(node, []byte{}, constants.IcmpProto, ip)
}

func RunPingTunnelCommand(c *cli.Context) {
	_nodeName := c.Args().Get(0)
	_gatewayIP := c.Args().Get(1)
	_tunnelIP := c.Args().Get(2)

	if _nodeName == "" || _gatewayIP == "" || _tunnelIP == "" {
		fmt.Println("Invalid command structure. Use 'run node ping tunnel <nodeName> <gatewayIP> <tunnelIP>'")
		return
	}

	node := (*Topology).GetNodeByName(_nodeName)
	if node == nil {
		fmt.Println("Node not found")
		return
	}
	emptyIPAddress := data.IPAddress{}
	ip := data.StringToIPAddress(_gatewayIP)
	if bytes.Equal(ip[:], emptyIPAddress[:]) {
		fmt.Println("Invalid IP address")
		return
	}
	tunnelIP := data.StringToIPAddress(_tunnelIP)
	if bytes.Equal(tunnelIP[:], emptyIPAddress[:]) {
		fmt.Println("Invalid IP address")
		return
	}

	ipHeader := &data.IPHeader{}
	ipHeader.Init()
	ipHeader.Protocol = constants.IcmpProto
	copy(ipHeader.DestinationIP[:], ip[:])
	copy(ipHeader.SourceIP[:], node.Properties.LB[:])
	ipHeader.IHL = uint8(unsafe.Sizeof(data.IPHeader{}) / 4)

	layers.PacketDemoteToLayer3(node, ipHeader.SerializeIPHeader(), constants.IpInIpProto, tunnelIP)
}

func ConfigNodeRoute(c *cli.Context) {
	_nodeName := c.Args().Get(0)
	_ipAddress := c.Args().Get(1)
	_mask := c.Args().Get(2)
	_gatewayIP := c.Args().Get(3)
	_interfaceName := c.Args().Get(4)

	if _nodeName == "" || _ipAddress == "" || _mask == "" || _gatewayIP == "" || _interfaceName == "" {
		fmt.Println("Invalid command structure. Use 'config node route <nodeName> <ipAddress> <mask> <gatewayIP> <interfaceName>'")
		return
	}
	node := (*Topology).GetNodeByName(_nodeName)
	if node == nil {
		fmt.Println("Node not found")
		return
	}
	emptyIPAddress := data.IPAddress{}
	ip := data.StringToIPAddress(_ipAddress)
	if bytes.Equal(ip[:], emptyIPAddress[:]) {
		fmt.Println("Invalid IP address")
		return
	}
	mask, err := strconv.Atoi(_mask)

	if err != nil || mask < 0 || mask > 128 {
		fmt.Println("Invalid mask")
		return
	}
	gatewayIP := data.StringToIPAddress(_gatewayIP)
	if bytes.Equal(gatewayIP[:], emptyIPAddress[:]) {
		fmt.Println("Invalid IP address")
		return
	}

	for _, intf := range node.Interfaces {
		if intf == nil {
			continue
		}
		if intf.Name.String() == _interfaceName {
			node.Properties.RoutingTable.AddRoute(ip, rune(mask), &gatewayIP, &intf.Name)
			return
		}
	}

	fmt.Println("Invalid interface name")
}
