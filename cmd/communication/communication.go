package communication

import (
	"fmt"
	"net"
	"sync"
	"tcpip/cmd/communication/receive"
	"tcpip/data"
)

var UDPPortNumber uint = 4000
var UDPPortNumberMutex sync.Mutex

// getNextUDPPortNumber returns the next available UDP port number and increments the global counter.
func getNextUDPPortNumber() uint {
	UDPPortNumberMutex.Lock()
	defer UDPPortNumberMutex.Unlock()
	UDPPortNumber++
	return UDPPortNumber
}

// initUDPSocket initializes the UDP socket for a node.
func initUDPSocket(node *data.Node) {
	if node.UDPConn != nil {
		return
	}

	node.UDPPortNumber = getNextUDPPortNumber()
	destIP := "127.0.0.1"
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", destIP, node.UDPPortNumber))
	if err != nil {
		fmt.Printf("Error resolving UDP address for node %s: %v\n", node.NodeName, err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Printf("Error creating UDP socket for node %s: %v\n", node.NodeName, err)
		return
	}

	node.UDPConn = conn
}

// StartPacketReceiverThread initializes UDP sockets for nodes in the graph and starts packet receiver threads.
func StartPacketReceiverThread(topology *data.Graph) {
	for dllNode := topology.Nodes.Next; dllNode != nil; dllNode = dllNode.Next {
		initUDPSocket(dllNode.DllToNode())
	}

	var wg sync.WaitGroup

	for dllNode := topology.Nodes.Next; dllNode != nil; dllNode = dllNode.Next {
		wg.Add(1)
		go func(node *data.Node) {
			defer wg.Done()

			for {
				buffer := data.PacketWithAux{}
				_, _, err := node.UDPConn.ReadFromUDP(buffer[:])

				if err != nil {
					fmt.Printf("Error: Failed to receive packet for node %s\n", node.NodeName)
					return
				}

				receive.PacketReceive(node, buffer)
			}
		}(dllNode.DllToNode())
	}
	wg.Wait()
}
