package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"unicode"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	snapshotLen int32 = 1024
	promiscuous       = false
	err         error
	timeout     = 100 * time.Millisecond
	handle      *pcap.Handle
	ethLayer    layers.Ethernet
	ipLayer     layers.IPv4
	tcpLayer    layers.TCP
)

//IsASCIIPrintable => {true if char is printable ; false otherwise}
func IsASCIIPrintable(r rune) bool {
	if r > unicode.MaxASCII || !unicode.IsPrint(r) {
		return false
	}
	return true
}

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

	// Set filter for only tcp traffic. Can also filter port #s with this
	var filter = "tcp"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	//go through all the packages being received
	for packet := range packetSource.Packets() {
		parser := gopacket.NewDecodingLayerParser(
			layers.LayerTypeEthernet,
			&ethLayer,
			&ipLayer,
			&tcpLayer,
		)

		//divide packet into layers
		foundLayerTypes := []gopacket.LayerType{}
		err := parser.DecodeLayers(packet.Data(), &foundLayerTypes)
		if err != nil {
			fmt.Println("Trouble decoding layers: ", err)
		}

		//go through packet layers
		for _, layerType := range foundLayerTypes {
			//print IP info
			if layerType == layers.LayerTypeIPv4 {
				fmt.Println("IPv4: ", ipLayer.SrcIP, "->", ipLayer.DstIP)
			}
			//print TCP info
			if layerType == layers.LayerTypeTCP {
				fmt.Println("TCP Port: ", tcpLayer.SrcPort, "->", tcpLayer.DstPort)
				fmt.Println("TCP SYN:", tcpLayer.SYN, " | ACK:", tcpLayer.ACK)

				//go through all bytes in the tcp payload
				for i, octet := range tcpLayer.Payload {
					//if character is printable, print it; else print a "."
					if IsASCIIPrintable(rune(octet)) {
						fmt.Print(string(octet), " ")
					} else {
						fmt.Print(". ")
					}
					//every 40 characters printed, print a newline
					if (i+1)%40 == 0 {
						fmt.Println()
					}
				}

				fmt.Print("\n-------------------------------\n")
			}
		}
	}
}
