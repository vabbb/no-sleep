package main

import (
	"fmt"
	"regexp"
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

type nodet struct {
	nodeID  string
	time    int64
	hasFlag bool
	data    string
	hex     []byte
}

type connt struct {
	connID           string
	srcIP, dstIP     string
	srcPort, dstPort uint16
	start, end       int64 // as is returned by time.Now().UnixNano()
	hasFlag          bool  // regex find for flag{...} pattern
	// some redundancy for faster processing
	nodes [][2]nodet // printable representation of the data
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
type uniStream struct {
	bothStrims *bothStreams // maps to my bidirectional twin.
	payloads   []streamPayload
	done       bool // if true, we've seen the last packet we're going to for this stream.
}

// bidi stores each unidirectional side of a bidirectional stream.
//
// When a new stream comes in, if we don't have an opposite stream, a bidi is
// created with 'a' set to the new stream.  If we DO have an opposite stream,
// 'b' is set to the new stream.
type bothStreams struct {
	key             key        // Key of the first stream, mostly for logging.
	a, b            *uniStream // the two bidirectional streams.
	firstPacketSeen int64
	lastPacketSeen  int64 // last time we saw a packet from either stream.
	lastSpeaker     bool  // true="last packet was from 'a'", false otherwise
}

// bidiFactory implements tcpassmebly.StreamFactory
type bidiFactory struct {
	// bidiMap maps keys to bidirectional stream pairs.
	bidiMap map[key]*bothStreams
}

func (factory *bidiFactory) New(netFlow, tcpFlow gopacket.Flow) tcpassembly.Stream {
	// log.Tracef("New stream %v:%v started", netFlow, tcpFlow)
	s := &uniStream{}

	// Find the bidi bidirectional struct for this stream, creating a new one if
	// one doesn't already exist in the map.
	k := key{netFlow, tcpFlow}
	bd := factory.bidiMap[k]

	if bd == nil {
		bd = &bothStreams{
			a:           s,
			key:         k,
			lastSpeaker: true,
		}
		log.Debugf("[%v] created first side of bidirectional stream", bd.key)
		// Register bidirectional with the reverse key, so the matching stream going
		// the other direction will find it.
		factory.bidiMap[key{netFlow.Reverse(), tcpFlow.Reverse()}] = bd
	} else {
		log.Debugf("[%v] found second side of bidirectional stream", bd.key)
		bd.b = s
		if factory.bidiMap[k].lastSpeaker == true {
			// last one to talk was a

		} else {
			// last one to talk was b

		}
		// Clear out the bidi we're using from the map, just in case.
		delete(factory.bidiMap, k)
	}
	s.bothStrims = bd
	return s
}

// Reassembled is called whenever new packet data is available for reading.
// Reassembly objects contain stream data IN ORDER.
func (s *uniStream) Reassembled(reassemblies []tcpassembly.Reassembly) {
	for _, reassembly := range reassemblies {
		strimPayload := streamPayload{
			size:   0,
			hasSYN: reassembly.Start,
			hasFIN: reassembly.End,
			time:   reassembly.Seen.UnixNano(),
			data:   reassembly.Bytes,
		}

		s.payloads = append(s.payloads, strimPayload)
	}
}

// emptyStream is used to finish bidi that only have one stream, in
// collectOldStreams.
var emptyStream = &uniStream{done: true}

// collectOldStreams finds any streams that haven't received a packet within
// 'timeout', and sets/finishes the 'b' stream inside them.  The 'a' stream may
// still receive packets after this.
func (factory *bidiFactory) collectOldStreams() {
	cutoff := time.Now().Add(-timeout).UnixNano()
	for k, bd := range factory.bidiMap {
		if bd.lastPacketSeen < cutoff {
			log.Debugf("[%v] timing out old stream", bd.key)
			bd.b = emptyStream         // stub out b with an empty stream.
			delete(factory.bidiMap, k) // remove it from our map.
			bd.maybeFinish()           // if b was the last stream we were waiting for, finish up.
		}
	}
}

// maybeFinish will wait until both directions are complete, then print out
// stats.
func (bd *bothStreams) maybeFinish() {
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
		/*UPLOAD FLOWT TO MONGO HERE*/
		// insertFlowtDoc(flowToUpload)
		log.Debugf("[%v] FINISHED", bd.key)
	}
}

// ReassemblyComplete is called when the TCP assembler
// believes a stream has finished.
func (s *uniStream) ReassemblyComplete() {
	// diffSecs := float64(s.end.Sub(s.start)) / float64(time.Second)
	//var flowToUpload flowt

	log.Debugf("Reassembly of stream %v:%v complete ", //- start:%v end:%v bytes:%v",
		s.bothStrims.key.net, s.bothStrims.key.transport) // s.start, s.end, s.bytes)

	temp := make([]byte, len(s.payloads[0].data))
	for i, octet := range s.payloads[0].data {
		//if character is printable, add it; else add a "."
		if IsASCIIPrintable(rune(octet)) {
			temp[i] = byte(octet)
		} else {
			temp[i] = 0x2e
		}
	}

	// flowToUpload := &connt{
	// 	srcIP:   s.bidi.key.net.Src().String(),
	// 	dstIP:   s.bidi.key.net.Dst().String(),
	// 	srcPort: bytesToUint16(s.bidi.key.transport.Src().Raw()),
	// 	dstPort: bytesToUint16(s.bidi.key.transport.Dst().Raw()),
	// 	hasSYN:  s.payloads[0].hasSYN,
	// 	hasFIN:  s.payloads[0].hasFIN,
	// 	start:   s.firstPacketSeen,
	// 	end:     s.lastPacketSeen,
	// 	size:    s.payloads[0].size,
	// 	hex:     s.payloads[0].data,
	// 	data:    string(temp),
	// }

	// // connID is made of the 2 pairs IP:PORT
	// // they are SORTED so the connID is the same both ways
	// connIDPieces := []string{flowToUpload.srcIP + ":" + strconv.Itoa(int(flowToUpload.srcPort)),
	// 	flowToUpload.dstIP + ":" + strconv.Itoa(int(flowToUpload.dstPort))}
	// sort.Strings(connIDPieces)
	// flowToUpload.connID = connIDPieces[0] + "<=>" + connIDPieces[1]

	// // look for flags
	// flowToUpload.hasFlag = isFlagPresent(flowToUpload.data)

	// /* Concatenate all unique info of the flow in order
	// to make a unique MD5 which will be the flow's id */
	// t := append(s.bidi.key.net.Src().Raw(), s.bidi.key.net.Dst().Raw()...)
	// t = append(t, s.bidi.key.transport.Src().Raw()...)
	// t = append(t, s.bidi.key.transport.Dst().Raw()...)
	// // create a temporary buffer to hold the time converted from int64 to []byte
	// bytetime := new(bytes.Buffer)
	// binary.Write(bytetime, binary.LittleEndian, flowToUpload.start)
	// t = append(t, bytetime.Bytes()...)
	// t = append(t, flowToUpload.hex...)
	// md5sum := md5.Sum(t)
	// flowToUpload.flowID = hex.EncodeToString(md5sum[:])

	log.Debug("key: ", s.bothStrims.key)
	log.Debug("payloads for a:")
	for _, paeload := range s.bothStrims.a.payloads {
		log.Debug("[", time.Unix(paeload.time/1000000000, paeload.time%1000000000), "]:", string(paeload.data))
		log.Debug("-------------------------------\n\n")
	}
	log.Debug("payloads for b:")
	for _, paeload := range s.bothStrims.b.payloads {
		log.Debug("[", time.Unix(paeload.time/1000000000, paeload.time%1000000000), "]:", string(paeload.data))
		log.Debug("-------------------------------")
	}

	s.done = true
	s.bothStrims.maybeFinish()

}
