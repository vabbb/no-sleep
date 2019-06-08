package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
)

var (
	err error

	device = flag.String("i", "", "Interface to get packets from")
	fname  = flag.String("r", "", "Filename to read from, overrides -i")
	//"tcp and dst port 80"
	filter        = flag.String("f", "tcp", "BPF filter for pcap")
	logAllPackets = flag.Bool("v", false, `
	Logs every packet in great detail`)
	bufferedPerConnection = flag.Int("connection_max_buffer", 0, `
	Max packets to buffer for a single connection before skipping over a gap 
	in data and continuing to stream the connection after the buffer. If zero 
	or less, this is infinite.`)
	bufferedTotal = flag.Int("total_max_buffer", 0, `
	Max packets to buffer total before skipping over gaps in connections and
	continuing to stream connection data.  If zero or less, this is infinite`)
	help = flag.Bool("help", false, "Shows this output")

	handle        *pcap.Handle
	snapshotLen   int32 = 65536
	promiscuous         = false
	flushDuration       = time.Second * 30

	byteCount int64

	ethLayer     layers.Ethernet
	ip4Layer     layers.IPv4
	ip6Layer     layers.IPv6
	tlsLayer     layers.TLS
	tcpLayer     layers.TCP
	payloadLayer gopacket.Payload

	parser = gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet,
		&ethLayer, &ip4Layer, &ip6Layer, &tlsLayer, &tcpLayer, &payloadLayer)
)

func main() {
	//parse command line arguments
	flag.Parse()
	if *help == true {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *device == "" && *fname == "" {
		fmt.Print("Usage:\n\tsudo timon -r [dump.pcap]\n\n")
		fmt.Print("Show help:\n\tsudo timom -help\n\n")
		os.Exit(1)
	}

	// Open device
	log.Printf("starting capture on interface %q", *device)
	handle, err = pcap.OpenLive(*device, snapshotLen, promiscuous, flushDuration/2)
	if err != nil {
		log.Fatal("error opening pcap handle: ", err)
	}
	defer handle.Close()

	// Set filter for only tcp traffic. Can also filter port numbers
	err = handle.SetBPFFilter(*filter)
	if err != nil {
		log.Fatal("error setting BPF filter: ", err)
	}

	// Set up assembly
	streamFactory := &tcpStreamFactory{}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)
	assembler.MaxBufferedPagesPerConnection = *bufferedPerConnection
	assembler.MaxBufferedPagesTotal = *bufferedTotal
	defer assembler.FlushAll()

	nextFlush := time.Now().Add(flushDuration / 2)

	decoded := make([]gopacket.LayerType, 0, 4)

	// infinite loop for reading packets
loop:
	for {
		// Check to see if we should flush the streams we have
		// that haven't seen any new data in a while.  Note we set a
		// timeout on our PCAP handle, so this should happen even if we
		// never see packet data.
		if time.Now().After(nextFlush) {
			// flushing all streams that haven't seen packets
			// in the last 2 minutes
			assembler.FlushOlderThan(time.Now().Add(flushDuration))
			nextFlush = time.Now().Add(flushDuration / 2)
		}

		// copy packet from kernel buffers with ReadPacketData
		data, ci, err := handle.ReadPacketData()

		if err != nil {
			log.Printf("error getting packet: %v", err)
			continue
		}

		err = parser.DecodeLayers(data, &decoded)
		if err != nil {
			log.Printf("error decoding packet: %v", err)
			continue
		}
		if *logAllPackets {
			log.Printf("decoded the following layers: %v", decoded)
		}
		byteCount += int64(len(data))
		// Find either the IPv4 or IPv6 address to use as our network
		// layer.
		foundNetLayer := false
		var netFlow gopacket.Flow
		for _, typ := range decoded {
			switch typ {
			case layers.LayerTypeIPv4:
				netFlow = ip4Layer.NetworkFlow()
				foundNetLayer = true
			case layers.LayerTypeIPv6:
				netFlow = ip6Layer.NetworkFlow()
				foundNetLayer = true
			case layers.LayerTypeTCP:
				if foundNetLayer {
					assembler.AssembleWithTimestamp(
						netFlow,
						&tcpLayer,
						ci.Timestamp,
					)
				} else {
					log.Println("could not find IPv4 layer, ignoring")
				}
				continue loop
			}
		}
		log.Println("could not find TCP layer")
	}
}
