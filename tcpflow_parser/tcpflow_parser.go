package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"
	"unicode"

	log "github.com/sirupsen/logrus"
)

var (
	err error

	flagRegex = flag.String("flag", "((?:flag|cci?t?1?9?){[ a-zA-Z0-9-_]*})",
		"Directory to look into for pcap files")
	debug = flag.Bool("debug", false, "If this is set, uses debug mode")
)

type nodet struct {
	connID           string
	nodeID           string
	srcIP, dstIP     string
	srcPort, dstPort int
	hasFlag          bool
	time             int64
	data             string
	hex              []byte
}

func stripIP(s string) string {
	first, _ := strconv.Atoi(s[0:3])
	secnd, _ := strconv.Atoi(s[4:7])
	third, _ := strconv.Atoi(s[8:11])
	forth, _ := strconv.Atoi(s[12:15])
	return fmt.Sprintf("%v.%v.%v.%v", first, secnd, third, forth)
}

// IsASCIIPrintable will return true if char is printable
// else it will return false
func IsASCIIPrintable(r rune) bool {
	if (r < unicode.MaxASCII && unicode.IsPrint(r)) || r == '\n' {
		return true
	}
	return false
}

func isFlagPresent(a string) bool {
	r, _ := regexp.Compile(*flagRegex)
	return r.MatchString(a)
}

func bytesToPrintable(ra []byte) []byte {
	temp := make([]byte, len(ra))
	for i, octet := range ra {
		//if character is printable, add it; else add a "."
		if IsASCIIPrintable(rune(octet)) {
			temp[i] = byte(octet)
		} else {
			temp[i] = 0x2e
		}
	}
	return temp
}

//setup mongodb server
func init() {
	// Parse command line arguments
	flag.Parse()

	// DEBUG MODE (?)
	if *debug == true {
		log.SetLevel(log.TraceLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	connectDB(url) // connected
	getCollectionsFromDB(client, dbName, connections)
	getCollectionsFromDB(client, dbName, nodes)
}

//this only works when piped with tcpflow
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	// var previousScanned *nodet
	// itstimetoupload := false

	for scanner.Scan() {
		idLine := scanner.Text()

		nodeToUpload := &nodet{
			srcIP: stripIP(idLine[:15]),
			dstIP: stripIP(idLine[22:37]),
			time:  time.Now().UnixNano(),
		}

		nodeToUpload.srcPort, _ = strconv.Atoi(idLine[16:21])
		nodeToUpload.dstPort, _ = strconv.Atoi(idLine[38:43])

		connIDPieces := []string{nodeToUpload.srcIP + ":" + strconv.Itoa(nodeToUpload.srcPort),
			nodeToUpload.dstIP + ":" + strconv.Itoa(nodeToUpload.dstPort)}
		sort.Strings(connIDPieces)
		nodeToUpload.connID = idLine[:len(idLine)-2]
		nodeToUpload.connID = connIDPieces[0] + "<=>" + connIDPieces[1]

		//here i have information if it is a new or an old nodet struct
		// if previousScanned == nil {
		// 	itstimetoupload = false
		// } else if previousScanned.connID == nodeToUpload.connID &&
		// 		previousScanned.srcIP == nodeToUpload.srcIP &&
		// 		previousScanned.dstIP == nodeToUpload.dstIP &&
		// 		previousScanned.srcPort == nodeToUpload.srcPort &&
		// 		previousScanned.dstPort == nodeToUpload.dstPort {
		// 			itstimetoupload = false
		// } else {
		// 	itstimetoupload = true
		// }

		runeArray := []rune{}
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				break
			}

			linenum := []rune{}
			lastWasSpace := false
			for i, c := range line {
				//skip tcpflow's line number
				if i < 5 {
					linenum = append(linenum, c)
					continue
				}
				if c == ' ' {
					if lastWasSpace == true {
						break
					}
					lastWasSpace = true
					continue
				}
				lastWasSpace = false
				runeArray = append(runeArray, c)
			}

			// readByte := line[6:7]
		}
		// fmt.Println("          runarr: " + string(runeArray))

		blob, _ := hex.DecodeString(string(runeArray))
		dataz := string(bytesToPrintable(blob))

		/* Concatenate all unique info of the flow in order
		to make a unique MD5 which will be the flow's id */
		t := append([]byte(nodeToUpload.srcIP), nodeToUpload.dstIP...)
		// create a temporary buffer to hold the time converted from int64 to []byte
		bytetime := new(bytes.Buffer)
		binary.Write(bytetime, binary.LittleEndian, nodeToUpload.time)
		t = append(t, bytetime.Bytes()...)
		t = append(t, nodeToUpload.hex...)
		md5sum := md5.Sum(t)
		nodeToUpload.nodeID = hex.EncodeToString(md5sum[:])

		nodeToUpload.data = dataz
		nodeToUpload.hex = blob
		nodeToUpload.hasFlag = isFlagPresent(nodeToUpload.data)

		log.Traceln("***************************")
		log.Traceln("connid:  " + nodeToUpload.connID)
		log.Traceln("nodeid:  " + nodeToUpload.nodeID)
		log.Traceln("src:     " + nodeToUpload.srcIP + ":" + strconv.Itoa(nodeToUpload.srcPort))
		log.Traceln("dst:     " + nodeToUpload.dstIP + ":" + strconv.Itoa(nodeToUpload.dstPort))
		log.Traceln("time:   ", (nodeToUpload.time))
		log.Traceln("hasFalg:", nodeToUpload.hasFlag)
		log.Traceln("data:    " + nodeToUpload.data)
		log.Traceln("***************************")

		/**upload to db*/

		// if itstimetoupload{
		// 	insertNodetDoc(previousScanned)
		// 	previousScanned = nodeToUpload
		// }

		insertNodetDoc(nodeToUpload)
	}

	if err := scanner.Err(); err != nil {
		log.Warning(err)
	}
}
