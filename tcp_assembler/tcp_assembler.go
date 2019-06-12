package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	log "github.com/sirupsen/logrus"
)

var (
	err error

	pcapDir    = flag.String("d", "pcaps", "Directory to look into for pcap files")
	archiveDir = flag.String("a", "archive", "Directory where to store archived "+
		"pcap files")

	logAllPackets         = flag.Bool("w", false, "Logs every packet in great detail")
	bufferedPerConnection = flag.Int("connection_max_buffer", 0, "Max packets to"+
		"buffer for a single connection before skipping over a gap in data and "+
		"continuing to stream the connection after the buffer.\nIf zero or less, "+
		"this is infinite.")
	bufferedTotal = flag.Int("total_max_buffer", 0, "Max packets to buffer total "+
		"before skipping over gaps in connections and continuing to stream "+
		"connection data.\nIf zero or less, this is infinite")

	nodebug = flag.Bool("nodebug", false, "If this is set, uses production mode")
	help    = flag.Bool("help", false, "Shows this output")

	fname string

	handle        *pcap.Handle
	snapshotLen   int32 = 65536
	promiscuous         = false
	flushDuration       = time.Minute * 4

	// byteCount int64

	ethLayer     layers.Ethernet
	ip4Layer     layers.IPv4
	ip6Layer     layers.IPv6
	tlsLayer     layers.TLS
	tcpLayer     layers.TCP
	payloadLayer gopacket.Payload

	parser = gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet,
		&ethLayer, &ip4Layer, &ip6Layer, &tlsLayer, &tcpLayer, &payloadLayer)
)

func oldestPcap() (response string, arr error) {
	/* pcapFiles is a map (dictionary), in which
	   the keys are the last modify time and the value is the file's path */
	pcapFiles := make(map[int64]string)

	// find .pcap file to analyze
	err := filepath.Walk(*pcapDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".pcap" {
			pcapFiles[info.ModTime().UnixNano()] = path
		}
		return nil
	})
	if err != nil {
		log.Fatal("Error looking through files: ", err)
	}
	if len(pcapFiles) == 0 {
		log.Warning("No .pcap files found")
		return "", errors.New("no pcaps found, dawg")
	}

	oldest := int64(^uint64(0) >> 1) // this means "MAX_INT64"
	for t := range pcapFiles {
		if t < oldest {
			oldest = t
		}
	}
	return pcapFiles[oldest], nil
}

// init happens before main
func init() {
	// DEBUG MODE (?)
	if *nodebug {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}

	//parse command line arguments
	flag.Parse()

	// Create archiveDir if it doesnt exist
	if _, err := os.Stat(*archiveDir); os.IsNotExist(err) {
		os.Mkdir(*archiveDir, 0755)
	}

	// If user asked for help...
	if *help == true {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Connection to mongoDB server
	connectDB(url)
	getCollectionsFromDB(client, dbName, connections)
	getCollectionsFromDB(client, dbName, flows)
}

func main() {
	// Set up assembly
	streamFactory := &tcpStreamFactory{}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)
	assembler.MaxBufferedPagesPerConnection = *bufferedPerConnection
	assembler.MaxBufferedPagesTotal = *bufferedTotal
	defer assembler.FlushAll()

	nextFlush := time.Now().Add(flushDuration / 2)

	decoded := make([]gopacket.LayerType, 0, 4)

	for {
		// loop until a .pcap file to analyze is found
		for {
			fname, err = oldestPcap()
			if err == nil {
				break
			}
			time.Sleep(time.Second * 10)
		}

		// Open file
		log.Infof("opening file %q", fname)
		handle, err = pcap.OpenOffline(fname)
		if err != nil {
			log.Fatal("error opening pcap handle: ", err)
		}

		// READ PACKETS FROM PCAP FILE
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
				if err.Error() == "EOF" {
					//go to next .pcap file (if it exists, else wait around)
					break
				}
				log.Tracef("error getting packet: %v", err)
				continue
			}

			err = parser.DecodeLayers(data, &decoded)
			if err != nil {
				log.Tracef("error decoding packet: %v", err)
				continue
			}
			if *logAllPackets {
				log.Tracef("decoded the following layers: %v", decoded)
			}
			// byteCount += int64(len(data))

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
						log.Trace("could not find IPv4 layer, ignoring")
					}
					continue
				}
			}
		}
		// close pcap file
		handle.Close()

		// Move analyzed pcap files to archive folder (if not in DEBUG MODE)
		if log.GetLevel() < log.DebugLevel {
			// move pcap file to archive folder
			splitboi := strings.Split(fname, "/")
			onlyFname := splitboi[len(splitboi)-1]
			err := os.Rename(fname, *archiveDir+"/"+onlyFname)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Debugf("We are in debug mode! Analysis will restart in 10s")
			time.Sleep(time.Second * 60)
		}
	}
}
