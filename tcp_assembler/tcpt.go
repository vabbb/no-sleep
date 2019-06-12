package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"regexp"
	"sort"
	"strconv"
	"time"
	"unicode"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	log "github.com/sirupsen/logrus"
)

// IsASCIIPrintable will return true if char is printable
// else it will return false
func IsASCIIPrintable(r rune) bool {
	if r > unicode.MaxASCII || !unicode.IsPrint(r) {
		return false
	}
	return true
}

func bytesToUint16(a []byte) uint16 {
	return uint16(a[0])<<8 + uint16(a[1])
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

type flowt struct {
	flowID           string
	connID           string
	srcIP, dstIP     string
	srcPort, dstPort uint16
	start, end       int64 // as is returned by time.Now().UnixNano()
	hasFlag          bool  // regex find for flag{...} pattern
	favorite         bool  /* defaults to false, can only be
	changed from the front-end*/
	hasSYN, hasFIN bool
	size           int64
	// some redundancy for faster processing
	data string // printable representation of the data
	hex  []byte // hex representation of the data
}

func isFlagPresent(a string) bool {
	r, _ := regexp.Compile("((?:flag|cci?t?1?9?){[ a-zA-Z0-9-_]*})")
	return r.MatchString(a)
}

// tcpStreamFactory implements tcpassembly.StreamFactory
type tcpStreamFactory struct{}

// tcpStream will handle tcp packets
type tcpStream struct {
	net, transport   gopacket.Flow
	bytes            int64
	payload          []byte
	start, end       int64
	sawStart, sawEnd bool
}

func (factory *tcpStreamFactory) New(net, tport gopacket.Flow) tcpassembly.Stream {
	log.Tracef("new stream %v:%v started", net, tport)
	t := &tcpStream{
		net:       net,
		transport: tport,
		start:     int64(^uint64(0) >> 1), // this means "MAX_INT64"
	}
	t.end = 0
	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return t
}

// Reassembled is called whenever new packet data is available for reading.
// Reassembly objects contain stream data IN ORDER.
func (s *tcpStream) Reassembled(reassemblies []tcpassembly.Reassembly) {
	for _, reassembly := range reassemblies {
		s.start = min(s.start, reassembly.Seen.UnixNano())
		s.end = max(s.end, reassembly.Seen.UnixNano())
		s.payload = append(s.payload, reassembly.Bytes...)
		s.bytes += int64(len(reassembly.Bytes))

		s.sawStart = s.sawStart || reassembly.Start
		s.sawEnd = s.sawEnd || reassembly.End
	}
}

// ReassemblyComplete is called when the TCP assembler
// believes a stream has finished.
func (s *tcpStream) ReassemblyComplete() {
	// diffSecs := float64(s.end.Sub(s.start)) / float64(time.Second)
	//var flowToUpload flowt

	log.Tracef("Reassembly of stream %v:%v complete ", //- start:%v end:%v bytes:%v",
		s.net, s.transport) // s.start, s.end, s.bytes)

	// ignore flows that contain no payload
	if s.bytes > 0 {
		temp := make([]byte, len(s.payload))
		for i, octet := range s.payload {
			//if character is printable, add it; else add a "."
			if IsASCIIPrintable(rune(octet)) {
				temp[i] = byte(octet)
			} else {
				temp[i] = 0x2e
			}
		}

		flowToUpload := &flowt{
			srcIP:    s.net.Src().String(),
			dstIP:    s.net.Dst().String(),
			srcPort:  bytesToUint16(s.transport.Src().Raw()),
			dstPort:  bytesToUint16(s.transport.Dst().Raw()),
			favorite: false,
			hasSYN:   s.sawStart,
			hasFIN:   s.sawEnd,
			start:    s.start,
			end:      s.end,
			size:     s.bytes,
			hex:      s.payload,
			data:     string(temp),
		}

		// connID is made of the 2 pairs IP:PORT
		// they are SORTED so the connID is the same both ways
		connIDPieces := []string{flowToUpload.srcIP + ":" + strconv.Itoa(int(flowToUpload.srcPort)),
			flowToUpload.dstIP + ":" + strconv.Itoa(int(flowToUpload.dstPort))}
		sort.Strings(connIDPieces)
		flowToUpload.connID = connIDPieces[0] + "<->" + connIDPieces[1]

		// look for flags
		flowToUpload.hasFlag = isFlagPresent(flowToUpload.data)

		/* Concatenate all unique info of the flow in order
		to make a unique MD5 which will be the flow's id */
		t := append(s.net.Src().Raw(), s.net.Dst().Raw()...)
		t = append(t, s.transport.Src().Raw()...)
		t = append(t, s.transport.Dst().Raw()...)
		// create a temporary buffer to hold the time converted from int64 to []byte
		bytetime := new(bytes.Buffer)
		binary.Write(bytetime, binary.LittleEndian, flowToUpload.start)
		t = append(t, bytetime.Bytes()...)
		t = append(t, flowToUpload.hex...)
		md5sum := md5.Sum(t)
		flowToUpload.flowID = hex.EncodeToString(md5sum[:])

		log.Traceln("flowID:", flowToUpload.flowID)
		log.Traceln("connID:", flowToUpload.connID)
		log.Traceln("srcIP:", flowToUpload.srcIP)
		log.Traceln("dstIP:", flowToUpload.dstIP)
		log.Traceln("srcPort:", flowToUpload.srcPort)
		log.Traceln("dstPort:", flowToUpload.dstPort)
		log.Traceln("hasFlag:", flowToUpload.hasFlag)
		log.Traceln("favorite:", flowToUpload.favorite)
		log.Traceln("hasSYN, hasFIN: ", flowToUpload.hasSYN, ", ", flowToUpload.hasFIN)
		log.Traceln("start:", time.Unix(0, flowToUpload.start))
		log.Traceln("end:", time.Unix(0, flowToUpload.end))
		log.Traceln("dataFlow.size:", flowToUpload.size)
		log.Traceln("dataFlow.data:", flowToUpload.data)
		log.Traceln("-------------------------------\n")

		/*UPLOAD FLOWT TO MONGO HERE*/
		insertFlowtDoc(flowToUpload)
	}
}
