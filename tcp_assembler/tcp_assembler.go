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

// timeout is the length of time to wait befor flushing connections and
// bidirectional stream pairs.
const timeout time.Duration = time.Minute * 5

var (
	err error

	pcapDir    = flag.String("d", "pcaps", "Directory to look into for pcap files")
	archiveDir = flag.String("a", "archive", "Directory where to store archived "+
		"pcap files")

	flagRegex = flag.String("regex", "((?:flag|cci?t?1?9?){[ a-zA-Z0-9-_]*})",
		"regex to grep flags")

	nowait  = flag.Bool("nowait", false, "dont wait for second pcap in dir")
	nodebug = flag.Bool("nodebug", false, "If this is set, uses production mode")
	help    = flag.Bool("help", false, "Shows this output")

	fname string

	handle      *pcap.Handle
	snapshotLen int32 = 65536
	promiscuous       = false

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
		return "", errors.New("No pcaps found, dawg")
	}
	if *nodebug == true {
		if *nowait == false {
			if len(pcapFiles) == 1 {
				log.Info("Only 1 .pcap file was found. Waiting for one more, " +
					"to be sure tcpdump has finished writing on it")
				return "", errors.New("Only 1 .pcap file, dawg")
			}
		}
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
	// Parse command line arguments
	flag.Parse()

	// DEBUG MODE (?)
	if *nodebug == true {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}

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
	streamFactory := &bidiFactory{bidiMap: make(map[key]*bidi)}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)
	// Limit memory usage by auto-flushing connection state if we get over 100K
	// packets in memory, or over 1000 for a single stream.
	assembler.MaxBufferedPagesTotal = 100000
	assembler.MaxBufferedPagesPerConnection = 1000
	defer assembler.FlushAll()

	nextFlush := time.Now().Add(timeout / 2)

	decoded := make([]gopacket.LayerType, 0, 4)

	// Look for pcap file. If not found, wait 5s and try again
	for {
		fname, err = oldestPcap()
		if err == nil {
			break
		}
		time.Sleep(time.Second * 5)
	}

	for {
		// Open file
		log.Infof("Analyzing file %q", fname)
		handle, err = pcap.OpenOffline(fname)
		timeFileWasOpen := time.Now()
		if err != nil {
			log.Fatal("Error opening pcap handle: ", err)
		}

		// READ PACKETS FROM PCAP FILE
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
					// File is done, we can exit the loop and close the handle
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
			time.Sleep(time.Second * 10)
		}

		// Loop until a valid .pcap file is found
		// timeToSleep is used to syncronize the sleeps with the last file open
		timeToSleep := timeFileWasOpen.Add(time.Second * 20)
		for i := 0; true; i++ {
			fname, err = oldestPcap()
			if err == nil {
				break
			}
			time.Sleep(time.Until(timeToSleep))
			timeToSleep = timeToSleep.Add(time.Second * 20)
			if i >= 3 {
				log.Warning("At least one minute has passed and no new .pcap " +
					"files were found!")
			}
		}

	}
}
