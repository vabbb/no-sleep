package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:\n\tsudo timon [interface]\n")
		return
	}

	var (
		device            = os.Args[1]
		snapshotLen int32 = 1024
		promiscuous       = false
		err         error
		timeout     time.Duration = 1 * time.Second
		handle      *pcap.Handle
	)

	// Open device
	handle, err = pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// Process packet here
		fmt.Println(packet)
	}
}
