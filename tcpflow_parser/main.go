package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"
	"unicode"
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

//setup mongodb server
func init() {
	connectDB(url) // connected
	getCollectionsFromDB(client, dbName, connections)
	getCollectionsFromDB(client, dbName, nodes)
}

// IsASCIIPrintable will return true if char is printable
// else it will return false
func IsASCIIPrintable(r rune) bool {
	if r > unicode.MaxASCII || !unicode.IsPrint(r) {
		return false
	}
	return true
}

func isFlagPresent(a string) bool {
	r, _ := regexp.Compile("((?:flag|cci?t?1?9?){[ a-zA-Z0-9-_]*})")
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

//this only works when piped with tcpflow
func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		idLine := scanner.Text()

		nodeToUpload := &nodet{
			srcIP: idLine[:15],
			dstIP: idLine[22:37],
			time:  time.Now().UnixNano(),
		}

		nodeToUpload.srcPort, _ = strconv.Atoi(idLine[16:21])
		nodeToUpload.dstPort, _ = strconv.Atoi(idLine[38:43])

		connIDPieces := []string{nodeToUpload.srcIP + ":" + strconv.Itoa(int(nodeToUpload.srcPort)),
			nodeToUpload.dstIP + ":" + strconv.Itoa(int(nodeToUpload.dstPort))}
		sort.Strings(connIDPieces)
		nodeToUpload.connID = idLine[:len(idLine)-2]
		nodeToUpload.connID = connIDPieces[0] + "<=>" + connIDPieces[1]

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

		fmt.Println("connid:  " + nodeToUpload.connID)
		fmt.Println("nodeid:  " + nodeToUpload.nodeID)
		fmt.Println("src:     "+nodeToUpload.srcIP+"  ", (nodeToUpload.srcPort))
		fmt.Println("dst:     "+nodeToUpload.dstIP+"  ", (nodeToUpload.dstPort))
		fmt.Println("time:   ", (nodeToUpload.time))
		fmt.Println("data:    " + nodeToUpload.data)
		fmt.Println("hasFalg:", nodeToUpload.hasFlag)
		fmt.Println("___________________________________-")

		/**upload to db*/
		insertNodetDoc(nodeToUpload)
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
