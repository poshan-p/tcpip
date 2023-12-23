package topology

import (
	"tcpip/constants"
	"tcpip/data"
	"tcpip/layers"
)

func CyclicTopology() *data.Graph {
	var topology = data.CreateGraph("Graph")

	var node1 = topology.CreateNode("node1")
	var node2 = topology.CreateNode("node2")
	var node3 = topology.CreateNode("node2")

	data.InsertLink(node1, node2, "eth0/0", "eth0/1", 1)
	data.InsertLink(node2, node3, "eth0/2", "eth0/3", 1)
	data.InsertLink(node1, node3, "eth0/4", "eth0/5", 1)

	node1.SetLbAddress(data.StringToIPAddress("122.1.1.0"))
	node1.SetIntfIPAddress("eth0/4", data.StringToIPAddress("40.1.1.1"), 24)
	node1.SetIntfIPAddress("eth0/0", data.StringToIPAddress("20.1.1.1"), 24)

	node2.SetLbAddress(data.StringToIPAddress("122.1.1.1"))
	node2.SetIntfIPAddress("eth0/1", data.StringToIPAddress("20.1.1.2"), 24)
	node2.SetIntfIPAddress("eth0/2", data.StringToIPAddress("30.1.1.1"), 24)

	node3.SetLbAddress(data.StringToIPAddress("122.1.1.2"))
	node3.SetIntfIPAddress("eth0/3", data.StringToIPAddress("30.1.1.2"), 24)
	node3.SetIntfIPAddress("eth0/5", data.StringToIPAddress("40.1.1.2"), 24)

	return topology
}

func LinearTopology() *data.Graph {
	var topology = data.CreateGraph("Graph")

	var node1 = topology.CreateNode("node1")
	var node2 = topology.CreateNode("node2")
	var node3 = topology.CreateNode("node2")

	data.InsertLink(node1, node2, "eth0/0", "eth0/1", 1)
	data.InsertLink(node2, node3, "eth0/2", "eth0/3", 1)

	node1.SetLbAddress(data.StringToIPAddress("122.1.1.0"))
	node1.SetIntfIPAddress("eth0/0", data.StringToIPAddress("20.1.1.1"), 24)

	node2.SetLbAddress(data.StringToIPAddress("122.1.1.1"))
	node2.SetIntfIPAddress("eth0/1", data.StringToIPAddress("20.1.1.2"), 24)
	node2.SetIntfIPAddress("eth0/2", data.StringToIPAddress("30.1.1.1"), 24)

	node3.SetLbAddress(data.StringToIPAddress("122.1.1.2"))
	node3.SetIntfIPAddress("eth0/3", data.StringToIPAddress("30.1.1.2"), 24)

	return topology
}

func SwitchTopology() *data.Graph {
	var topology = data.CreateGraph("Dual switch topology")
	var H1 = topology.CreateNode("H1")
	H1.SetLbAddress(data.StringToIPAddress("122.1.1.1"))
	var H2 = topology.CreateNode("H2")
	H2.SetLbAddress(data.StringToIPAddress("122.1.1.2"))
	var H3 = topology.CreateNode("H3")
	H3.SetLbAddress(data.StringToIPAddress("122.1.1.3"))
	var H4 = topology.CreateNode("H4")
	H4.SetLbAddress(data.StringToIPAddress("122.1.1.4"))
	var H5 = topology.CreateNode("H5")
	H5.SetLbAddress(data.StringToIPAddress("122.1.1.5"))
	var H6 = topology.CreateNode("H6")
	H6.SetLbAddress(data.StringToIPAddress("122.1.1.6"))

	var L2SW1 = topology.CreateNode("L2SW1")
	var L2SW2 = topology.CreateNode("L2SW2")

	data.InsertLink(H1, L2SW1, "eth0/1", "eth0/2", 1)
	data.InsertLink(H2, L2SW1, "eth0/3", "eth0/7", 1)
	data.InsertLink(H3, L2SW1, "eth0/4", "eth0/6", 1)
	data.InsertLink(L2SW1, L2SW2, "eth0/5", "eth0/7", 1)
	data.InsertLink(H5, L2SW2, "eth0/8", "eth0/9", 1)
	data.InsertLink(H4, L2SW2, "eth0/11", "eth0/12", 1)
	data.InsertLink(H6, L2SW2, "eth0/11", "eth0/10", 1)

	H1.SetIntfIPAddress("eth0/1", data.StringToIPAddress("10.1.1.1"), 24)
	H2.SetIntfIPAddress("eth0/3", data.StringToIPAddress("10.1.1.2"), 24)
	H3.SetIntfIPAddress("eth0/4", data.StringToIPAddress("10.1.1.3"), 24)
	H4.SetIntfIPAddress("eth0/11", data.StringToIPAddress("10.1.1.4"), 24)
	H5.SetIntfIPAddress("eth0/8", data.StringToIPAddress("10.1.1.5"), 24)
	H6.SetIntfIPAddress("eth0/11", data.StringToIPAddress("10.1.1.6"), 24)

	layers.SetIntfL2Mode(L2SW1, "eth0/2", constants.ACCESS)
	layers.SetIntfVLAN(L2SW1, "eth0/2", 10)
	layers.SetIntfL2Mode(L2SW1, "eth0/7", constants.ACCESS)
	layers.SetIntfVLAN(L2SW1, "eth0/7", 10)
	layers.SetIntfL2Mode(L2SW1, "eth0/5", constants.TRUNK)
	layers.SetIntfVLAN(L2SW1, "eth0/5", 10)
	layers.SetIntfVLAN(L2SW1, "eth0/5", 11)
	layers.SetIntfL2Mode(L2SW1, "eth0/6", constants.ACCESS)
	layers.SetIntfVLAN(L2SW1, "eth0/6", 11)

	layers.SetIntfL2Mode(L2SW2, "eth0/7", constants.TRUNK)
	layers.SetIntfVLAN(L2SW2, "eth0/7", 10)
	layers.SetIntfVLAN(L2SW2, "eth0/7", 11)
	layers.SetIntfL2Mode(L2SW2, "eth0/9", constants.ACCESS)
	layers.SetIntfVLAN(L2SW2, "eth0/9", 10)
	layers.SetIntfL2Mode(L2SW2, "eth0/10", constants.ACCESS)
	layers.SetIntfVLAN(L2SW2, "eth0/10", 10)
	layers.SetIntfL2Mode(L2SW2, "eth0/12", constants.ACCESS)
	layers.SetIntfVLAN(L2SW2, "eth0/12", 11)

	return topology
}

func LinearTopologyRouting() *data.Graph {
	var topology = data.CreateGraph("3 node linear topology")
	var R1 = topology.CreateNode("R1")
	var R2 = topology.CreateNode("R2")
	var R3 = topology.CreateNode("R3")

	data.InsertLink(R1, R2, "eth0/1", "eth0/2", 1)
	data.InsertLink(R2, R3, "eth0/3", "eth0/4", 1)

	R1.SetLbAddress(data.StringToIPAddress("122.1.1.1"))
	R1.SetIntfIPAddress("eth0/1", data.StringToIPAddress("10.1.1.1"), 24)

	R2.SetLbAddress(data.StringToIPAddress("122.1.1.2"))
	R2.SetIntfIPAddress("eth0/2", data.StringToIPAddress("10.1.1.2"), 24)
	R2.SetIntfIPAddress("eth0/3", data.StringToIPAddress("11.1.1.2"), 24)

	R3.SetLbAddress(data.StringToIPAddress("122.1.1.3"))
	R3.SetIntfIPAddress("eth0/4", data.StringToIPAddress("11.1.1.1"), 24)

	return topology
}

func SquareTopology() *data.Graph {
	topology := data.CreateGraph("Square Topology")
	R1 := topology.CreateNode("R1")
	R2 := topology.CreateNode("R2")
	R3 := topology.CreateNode("R3")
	R4 := topology.CreateNode("R4")

	data.InsertLink(R1, R2, "eth0/0", "eth0/1", 1)
	data.InsertLink(R2, R3, "eth0/2", "eth0/3", 1)
	data.InsertLink(R3, R4, "eth0/4", "eth0/5", 1)
	data.InsertLink(R4, R1, "eth0/6", "eth0/7", 1)

	R1.SetLbAddress(data.StringToIPAddress("122.1.1.1"))
	R1.SetIntfIPAddress("eth0/0", data.StringToIPAddress("10.1.1.1"), 24)
	R1.SetIntfIPAddress("eth0/7", data.StringToIPAddress("40.1.1.2"), 24)

	R2.SetLbAddress(data.StringToIPAddress("122.1.1.2"))
	R2.SetIntfIPAddress("eth0/1", data.StringToIPAddress("10.1.1.2"), 24)
	R2.SetIntfIPAddress("eth0/2", data.StringToIPAddress("20.1.1.1"), 24)

	R3.SetLbAddress(data.StringToIPAddress("122.1.1.3"))
	R3.SetIntfIPAddress("eth0/3", data.StringToIPAddress("20.1.1.2"), 24)
	R3.SetIntfIPAddress("eth0/4", data.StringToIPAddress("30.1.1.1"), 24)

	R4.SetLbAddress(data.StringToIPAddress("122.1.1.4"))
	R4.SetIntfIPAddress("eth0/5", data.StringToIPAddress("30.1.1.2"), 24)
	R4.SetIntfIPAddress("eth0/6", data.StringToIPAddress("40.1.1.1"), 24)

	return topology
}
