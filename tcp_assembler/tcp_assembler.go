package main

import (
	"flag"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	log "github.com/sirupsen/logrus"
)

// timeout is the length of time to wait befor flushing connections and
// bidirectional stream pairs.
const timeout time.Duration = time.Minute * 5

var (
	err error

	readFrom = flag.String("r", "", "Directory to look into for pcap files")
	iface    = flag.String("i", "", "Interface to monitor")
	filter   = flag.String("filter", "", "BPF filter")
	//"((?:flag|cci?t?1?9?){[ a-zA-Z0-9-_]*})"
	flagRegex = flag.String("regex", "([A-Z0-9]{31}=)", "Regex to grep flags")

	debug = flag.Bool("debug", false, "If this is set, uses production mode")
	help  = flag.Bool("help", false, "Shows this output")

	file = &os.File{}

	handle *pcap.Handle

	ethLayer     layers.Ethernet
	ip4Layer     layers.IPv4
	ip6Layer     layers.IPv6
	tlsLayer     layers.TLS
	tcpLayer     layers.TCP
	payloadLayer gopacket.Payload

	parser = gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet,
		&ethLayer, &ip4Layer, &ip6Layer, &tlsLayer, &tcpLayer, &payloadLayer)
)

// init happens before main
func init() {
	// Parse command line arguments
	flag.Parse()

	// DEBUG MODE (?)
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// If user asked for help...
	if *help == true {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Connection to mongoDB server
	connectDB(url)
	getCollectionsFromDB(client, dbName, flows)
}

func main() {
	// Set up assembly
	streamFactory := &bidiFactory{bidiMap: make(map[key]*bothStreams)}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)
	// Limit memory usage by auto-flushing connection state if we get over 100K
	// packets in memory, or over 1000 for a single stream.
	assembler.MaxBufferedPagesTotal = 0
	assembler.MaxBufferedPagesPerConnection = 0
	defer assembler.FlushAll()

	nextFlush := time.Now().Add(timeout / 2)

	decoded := make([]gopacket.LayerType, 0, 4)

	if *iface != "" {
		// Open file
		log.Infof("Opening interface %q", *iface)
		handle, err = pcap.OpenLive(*iface, int32(262144), false, pcap.BlockForever)
		if err != nil {
			log.Fatal("Error opening pcap handle: ", err)
		}
		if err := handle.SetBPFFilter(*filter); err != nil {
			log.Fatal("error setting BPF filter: ", err)
		}

	} else if *readFrom != "" {
		if *readFrom == "-" {
			file = os.Stdin
		} else {
			file, err = os.Open(*readFrom)
		}

		// Open file
		log.Infof("Analyzing file %v", file.Name())
		handle, err = pcap.OpenOfflineFile(file)
		if err != nil {
			log.Fatal("Error opening pcap handle: ", err)
		}
	} else {
		flag.PrintDefaults()
		log.Fatal("Specify where you want to capture from!")
	}
	// READ PACKETS
	for {
		/* Check to see if we should flush the streams we have that
		   haven't seen any new data in a while.  Note we set a timeout
		   on our PCAP handle, so this should happen even if we never
		   see packet data. */
		if time.Now().After(nextFlush) {
			// flushing all streams that haven't seen packets
			// in the last 2 minutes
			assembler.FlushOlderThan(time.Now().Add(timeout))
			nextFlush = time.Now().Add(timeout / 2)
		}

		// copy packet from kernel buffers with ReadPacketData
		data, ci, err := handle.ReadPacketData()

		if err != nil {
			if err.Error() == "EOF" {
				// File reached EOF, we can exit the loop and close the handle
				break
			}
			log.Tracef("Error getting packet: %v", err)
			continue
		}

		err = parser.DecodeLayers(data, &decoded)
		if err != nil {
			log.Tracef("Error decoding packet: %v", err)
			continue
		}

		// Find either the IPv4 or IPv6 address to use as our network layer.
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
					log.Trace("Could not find IPv4 layer, ignoring")
				}
				continue
			}
		}
	}
	// Close pcap file
	handle.Close()
}
