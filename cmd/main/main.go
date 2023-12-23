package main

import (
	"tcpip/cmd"
	"tcpip/cmd/communication"
)

func main() {
	go communication.StartPacketReceiverThread(cmd.Topology)
	cmd.Init()
}
