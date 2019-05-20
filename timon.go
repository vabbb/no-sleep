package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var (
	snapshotLen int32 = 1024
	promiscuous       = false
	err         error
	timeout     = 100 * time.Millisecond
	handle      *pcap.Handle
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print("Usage:\n\tsudo timon [interface]\n\n")
		return
	}

	// 1st user given attribute is the interface to monitor
	var device = os.Args[1]

	// Open device
	handle, err = pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Set filter
	var filter = "tcp"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Only capturing TCP.")

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// Process packet here
		fmt.Println(packet)
	}
}
