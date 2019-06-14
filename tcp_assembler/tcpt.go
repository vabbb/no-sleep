package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
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
	hasSYN, hasFIN   bool
	size             int64
	// some redundancy for faster processing
	data string // printable representation of the data
	hex  []byte // hex representation of the data
}

func isFlagPresent(a string) bool {
	r, _ := regexp.Compile(*flagRegex)
	return r.MatchString(a)
}

// tcpStreamFactory implements tcpassembly.StreamFactory
type bidiStreamFactory struct{}

// key is used to map bidirectional streams to each other.
type key struct {
	net, transport gopacket.Flow
}

// String prints out the key in a human-readable fashion.
func (k key) String() string {
	return fmt.Sprintf("%v:%v", k.net, k.transport)
}

type streamPayload struct {
	size           int64
	data           []byte
	hasSYN, hasFIN bool
	time           int64
}

// bidiStream will handle tcp packets
type bidiStream struct {
	bidi       *bidi // maps to my bidirectional twin.
	payloads   []streamPayload
	start, end int64
	done       bool // if true, we've seen the last packet we're going to for this stream.
}

// bidi stores each unidirectional side of a bidirectional stream.
//
// When a new stream comes in, if we don't have an opposite stream, a bidi is
// created with 'a' set to the new stream.  If we DO have an opposite stream,
// 'b' is set to the new stream.
type bidi struct {
	key            key         // Key of the first stream, mostly for logging.
	a, b           *bidiStream // the two bidirectional streams.
	lastPacketSeen time.Time   // last time we saw a packet from either stream.
}

// bidiFactory implements tcpassmebly.StreamFactory
type bidiFactory struct {
	// bidiMap maps keys to bidirectional stream pairs.
	bidiMap map[key]*bidi
}

func (factory *bidiFactory) New(netFlow, tcpFlow gopacket.Flow) tcpassembly.Stream {
	// log.Tracef("New stream %v:%v started", net, transport)
	s := &bidiStream{}

	// Find the bidi bidirectional struct for this stream, creating a new one if
	// one doesn't already exist in the map.
	k := key{netFlow, tcpFlow}
	bd := factory.bidiMap[k]

	if bd == nil {
		bd = &bidi{a: s, key: k}
		log.Debugf("[%v] created first side of bidirectional stream", bd.key)
		// Register bidirectional with the reverse key, so the matching stream going
		// the other direction will find it.
		factory.bidiMap[key{netFlow.Reverse(), tcpFlow.Reverse()}] = bd
	} else {
		log.Debugf("[%v] found second side of bidirectional stream", bd.key)
		bd.b = s
		// Clear out the bidi we're using from the map, just in case.
		delete(factory.bidiMap, k)
	}
	s.bidi = bd
	return s
}

// Reassembled is called whenever new packet data is available for reading.
// Reassembly objects contain stream data IN ORDER.
func (s *bidiStream) Reassembled(reassemblies []tcpassembly.Reassembly) {
	for _, reassembly := range reassemblies {
		strimPayload := streamPayload{
			size:   0,
			hasSYN: reassembly.Start,
			hasFIN: reassembly.End,
			time:   reassembly.Seen.UnixNano(),
			data:   reassembly.Bytes,
		}

		s.payloads = append(s.payloads, strimPayload)

		s.start = min(s.start, reassembly.Seen.UnixNano())
		s.end = max(s.end, reassembly.Seen.UnixNano())
	}
}

// emptyStream is used to finish bidi that only have one stream, in
// collectOldStreams.
var emptyStream = &bidiStream{done: true}

// collectOldStreams finds any streams that haven't received a packet within
// 'timeout', and sets/finishes the 'b' stream inside them.  The 'a' stream may
// still receive packets after this.
func (factory *bidiFactory) collectOldStreams() {
	cutoff := time.Now().Add(-timeout)
	for k, bd := range factory.bidiMap {
		if bd.lastPacketSeen.Before(cutoff) {
			log.Debugf("[%v] timing out old stream", bd.key)
			bd.b = emptyStream         // stub out b with an empty stream.
			delete(factory.bidiMap, k) // remove it from our map.
			bd.maybeFinish()           // if b was the last stream we were waiting for, finish up.
		}
	}
}

// maybeFinish will wait until both directions are complete, then print out
// stats.
func (bd *bidi) maybeFinish() {
	switch {
	case bd.a == nil:
		log.Fatalf("[%v] a should always be non-nil, since it's set when bidis are created", bd.key)
	case !bd.a.done:
		log.Debugf("[%v] still waiting on first stream", bd.key)
	case bd.b == nil:
		log.Debugf("[%v] no second stream yet", bd.key)
	case !bd.b.done:
		log.Debugf("[%v] still waiting on second stream", bd.key)
	default:
		log.Debugf("[%v] FINISHED", bd.key)
		// log.Debugf("[%v] FINISHED, bytes: %d tx, %d rx", bd.key,
		// 	bd.a.bytes, bd.b.bytes)
	}
}

// ReassemblyComplete is called when the TCP assembler
// believes a stream has finished.
func (s *bidiStream) ReassemblyComplete() {
	// diffSecs := float64(s.end.Sub(s.start)) / float64(time.Second)
	//var flowToUpload flowt

	log.Debugf("Reassembly of stream %v:%v complete ", //- start:%v end:%v bytes:%v",
		s.bidi.key.net, s.bidi.key.transport) // s.start, s.end, s.bytes)

	s.done = true
	s.bidi.maybeFinish()

	temp := make([]byte, len(s.payloads[0].data))
	for i, octet := range s.payloads[0].data {
		//if character is printable, add it; else add a "."
		if IsASCIIPrintable(rune(octet)) {
			temp[i] = byte(octet)
		} else {
			temp[i] = 0x2e
		}
	}

	flowToUpload := &flowt{
		srcIP:   s.bidi.key.net.Src().String(),
		dstIP:   s.bidi.key.net.Dst().String(),
		srcPort: bytesToUint16(s.bidi.key.transport.Src().Raw()),
		dstPort: bytesToUint16(s.bidi.key.transport.Dst().Raw()),
		hasSYN:  s.payloads[0].hasSYN,
		hasFIN:  s.payloads[0].hasFIN,
		start:   s.start,
		end:     s.end,
		size:    s.payloads[0].size,
		hex:     s.payloads[0].data,
		data:    string(temp),
	}

	// connID is made of the 2 pairs IP:PORT
	// they are SORTED so the connID is the same both ways
	connIDPieces := []string{flowToUpload.srcIP + ":" + strconv.Itoa(int(flowToUpload.srcPort)),
		flowToUpload.dstIP + ":" + strconv.Itoa(int(flowToUpload.dstPort))}
	sort.Strings(connIDPieces)
	flowToUpload.connID = connIDPieces[0] + "<=>" + connIDPieces[1]

	// look for flags
	flowToUpload.hasFlag = isFlagPresent(flowToUpload.data)

	/* Concatenate all unique info of the flow in order
	to make a unique MD5 which will be the flow's id */
	t := append(s.bidi.key.net.Src().Raw(), s.bidi.key.net.Dst().Raw()...)
	t = append(t, s.bidi.key.transport.Src().Raw()...)
	t = append(t, s.bidi.key.transport.Dst().Raw()...)
	// create a temporary buffer to hold the time converted from int64 to []byte
	bytetime := new(bytes.Buffer)
	binary.Write(bytetime, binary.LittleEndian, flowToUpload.start)
	t = append(t, bytetime.Bytes()...)
	t = append(t, flowToUpload.hex...)
	md5sum := md5.Sum(t)
	flowToUpload.flowID = hex.EncodeToString(md5sum[:])

	log.Debug("flowID: ", flowToUpload.flowID)
	log.Debug("connID: ", flowToUpload.connID)
	log.Debug("srcIP: ", flowToUpload.srcIP)
	log.Debug("dstIP: ", flowToUpload.dstIP)
	log.Debug("srcPort: ", flowToUpload.srcPort)
	log.Debug("dstPort: ", flowToUpload.dstPort)
	log.Debug("hasFlag: ", flowToUpload.hasFlag)
	log.Debug("hasSYN, hasFIN: ", flowToUpload.hasSYN, ", ", flowToUpload.hasFIN)
	log.Debug("start: ", time.Unix(0, flowToUpload.start))
	log.Debug("end: ", time.Unix(0, flowToUpload.end))
	log.Debug("dataFlow.size: ", flowToUpload.size)
	log.Debug("dataFlow.data: ", flowToUpload.data)
	log.Debug("-------------------------------\n\n")

	/*UPLOAD FLOWT TO MONGO HERE*/
	// insertFlowtDoc(flowToUpload)

}
