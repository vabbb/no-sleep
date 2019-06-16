package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	log "github.com/sirupsen/logrus"
)

// IsASCIIPrintable returns true if char is printable, else returns false
func IsASCIIPrintable(r rune) bool {
	if r < unicode.MaxASCII && (unicode.IsGraphic(r) || unicode.IsSpace(r)) {
		return true
	}
	return false
}

// Transform unprintable bytes into dots, then transtorm array into string
func toPrintable(a []byte) string {
	temp := make([]byte, len(a))
	for i, octet := range a {
		//if character is printable, add it; else add a "."
		if IsASCIIPrintable(rune(octet)) {
			temp[i] = byte(octet)
		} else {
			temp[i] = 0x2e
		}
	}
	return string(temp)
}

func bytesToUint16(a []byte) uint16 {
	return uint16(a[0])<<8 + uint16(a[1])
}

type flowt struct {
	flowID           string
	srcIP, dstIP     string
	srcPort, dstPort uint16
	start, end       int64 // as is returned by time.Now().UnixNano()
	hasFlag          bool  // regex find for flag{...} pattern
	seenSYN, seenFIN bool
	trafficSize      int
	// some redundancy for faster processing
	nodes []nodet // printable representation of the data
}

func (f flowt) String() string {
	r := ""
	r += "flowID: " + f.flowID
	if f.hasFlag {
		r += "\nHAS FLAG"
	}
	if f.seenSYN {
		r += "\nstarted with SYN"
	}
	if f.seenFIN {
		r += "\nended with FIN"
	}
	r += "\nTRAFFIC: " + strconv.Itoa(f.trafficSize) + " Bytes"
	r += "\nNODES :"
	for _, node := range f.nodes {

		r += "\n[" + node.srcIP + ":" + strconv.Itoa(int(node.srcPort))
		r += "->" + node.dstIP + ":" + strconv.Itoa(int(node.dstPort))
		r += "]["
		r += time.Unix(node.time/1000000000, node.time%1000000000).String()
		r += "]" + "[SIZE: " + strconv.Itoa(node.size) + "]"
		if node.isSrc {
			r += "[CLIENT]"
		} else {
			r += "[SERVER]"
		}
		r += ":\n" + string(node.printableData)
		r += ("\n")
	}

	return r
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

type nodet struct {
	time             int64
	srcIP, dstIP     string
	srcPort, dstPort uint16
	isSrc            bool
	hasSYN, hasFIN   bool
	hasFlag          bool
	size             int
	blob             []byte
	printableData    string
}

// bidiStream will handle tcp packets
type uniStream struct {
	net, transport gopacket.Flow
	bothStrims     *bothStreams // maps to my bidirectional twin.
	nodets         []nodet      // single nodes for this stream
	done           bool         // if true, we've seen the last packet we're going to for this stream.
}

// bidi stores each unidirectional side of a bidirectional stream.
//
// When a new stream comes in, if we don't have an opposite stream, a bidi is
// created with 'a' set to the new stream.  If we DO have an opposite stream,
// 'b' is set to the new stream.
type bothStreams struct {
	key            key        // Key of the first stream, mostly for logging.
	a, b           *uniStream // the two bidirectional streams.
	lastPacketSeen int64      // last time we saw a packet from either stream.
}

// bidiFactory implements tcpassmebly.StreamFactory
type bidiFactory struct {
	// bidiMap maps keys to bidirectional stream pairs.
	bidiMap map[key]*bothStreams
}

func (factory *bidiFactory) New(netFlow, tcpFlow gopacket.Flow) tcpassembly.Stream {
	// log.Tracef("New stream %v:%v started", netFlow, tcpFlow)
	s := &uniStream{
		net:       netFlow,
		transport: tcpFlow,
	}

	// Find the bidi bidirectional struct for this stream, creating a new one if
	// one doesn't already exist in the map.
	k := key{netFlow, tcpFlow}
	bd := factory.bidiMap[k]

	if bd == nil {
		bd = &bothStreams{
			a:   s,
			key: k,
		}
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
	s.bothStrims = bd
	return s
}

// Reassembled is called whenever new packet data is available for reading.
// Reassembly objects contain stream data IN ORDER.
func (s *uniStream) Reassembled(reassemblies []tcpassembly.Reassembly) {
	for _, reassembly := range reassemblies {
		reassNodet := nodet{
			srcIP:   s.net.Src().String(),
			dstIP:   s.net.Dst().String(),
			srcPort: bytesToUint16(s.transport.Src().Raw()),
			dstPort: bytesToUint16(s.transport.Dst().Raw()),
			hasSYN:  reassembly.Start,
			hasFIN:  reassembly.End,
			size:    0, // assigned in mergeAdjacentAndCalcTraffic, for optimization reasons
			time:    reassembly.Seen.UnixNano(),
			blob:    reassembly.Bytes,
		}

		// generate printableData
		reassNodet.printableData = toPrintable(reassNodet.blob)

		// append this proto-nodet to the array of nodets for this stream
		s.nodets = append(s.nodets, reassNodet)
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

// merge nodets belonging to two streams (they are already ordered by time)
func mergeSort(asp []nodet, bsp []nodet) []nodet {
	r := make([]nodet, len(asp)+len(bsp))
	for i, j := 0, 0; i < len(asp) && j < len(bsp); {
		if asp[i].time < bsp[j].time {
			r[i+j] = asp[i]
			i++
		} else {
			r[i+j] = bsp[j]
			j++
		}
	}
	return r
}

// merge 2 nodets into a single nodet
func (n *nodet) merge(a *nodet) *nodet {
	(*n).hasSYN = n.hasSYN || a.hasSYN
	(*n).hasFIN = n.hasFIN || a.hasFIN
	(*n).printableData += a.printableData
	(*n).size += a.size
	(*n).isSrc = a.isSrc
	(*n).blob = append(n.blob, a.blob...)
	return n
}

// merge adjacent nodes if both have same endpoints and direction
// also calulate total traffic bytes <- this is an optimization to only have
// one for loop that does it all
func mergeAdjacentAndCalcTraffic(a []nodet) ([]nodet, int) {
	traffic := 0
	if len(a) == 0 {
		return []nodet{}, traffic
	}
	r := []nodet{
		a[0],
	}
	a[0].size = len(a[0].blob)
	r[0].size = len(r[0].blob)
	r[0] = a[0]
	traffic += a[0].size
	i, j := 1, 1
	for i < len(a) {
		a[i].size = len(a[i].blob)
		traffic += a[i].size

		// if same dest and same src
		if a[i].srcPort == a[i-1].srcPort && a[i].dstPort == a[i-1].dstPort &&
			a[i].srcIP == a[i-1].srcIP && a[i].dstIP == a[i-1].dstIP {
			r[j-1] = *r[j-1].merge(&a[i])
		} else {
			r = append(r, a[i])
			j++
		}
		i++
	}
	return r, traffic
}

func (n *nodet) checkIfSource(f *flowt) bool {
	if f.srcIP == n.srcIP && f.srcPort == n.srcPort &&
		f.dstIP == n.dstIP && f.dstPort == n.dstPort {
		return true
	}
	return false
}

// remove SYN, SYN/ACK and FIN packets
func transferSYNsAndFINsToFlowt(a []nodet, f *flowt) []nodet {
	r := []nodet{}
	for i := 0; i < len(a)-1; i++ {
		if a[i].hasSYN {
			f.seenSYN = true
		}
		if a[i].hasFIN {
			f.seenFIN = true
		}
		if len(a[i].blob) > 0 {
			// for optimization reasons, we check here if packet is from src or dst
			a[i].isSrc = a[i].checkIfSource(f)
			r = append(r, a[i])
		}
	}
	return r
}

// maybeFinish will wait until both directions are complete, then print out
// stats.
func (both *bothStreams) maybeFinish() {
	switch {
	case both.a == nil:
		log.Fatalf("[%v] a should always be non-nil, since it's set when bidis are created", both.key)
	case !both.a.done:
		log.Debugf("[%v] still waiting on first stream", both.key)
	case both.b == nil:
		log.Debugf("[%v] no second stream yet", both.key)
	case !both.b.done:
		log.Debugf("[%v] still waiting on second stream", both.key)
	default:
		log.Debugf("[%v] FINISHED", both.key)
		flowToUpload := &flowt{
			srcIP:   both.key.net.Src().String(),
			dstIP:   both.key.net.Dst().String(),
			srcPort: bytesToUint16(both.key.transport.Src().Raw()),
			dstPort: bytesToUint16(both.key.transport.Dst().Raw()),
		}

		flowToUpload.flowID = flowToUpload.srcIP + ":" + strconv.Itoa(int(flowToUpload.srcPort)) +
			" => " + flowToUpload.dstIP + ":" + strconv.Itoa(int(flowToUpload.dstPort))

		temp0 := transferSYNsAndFINsToFlowt(
			mergeSort(both.a.nodets, both.b.nodets),
			flowToUpload,
		)

		flowToUpload.nodes, flowToUpload.trafficSize =
			mergeAdjacentAndCalcTraffic(temp0)

		// fill hasFlag fields for each nodet, and for flowt too
		for _, node := range flowToUpload.nodes {
			if isFlagPresent(node.printableData) == true {
				node.hasFlag = true
				flowToUpload.hasFlag = true
			} else {
				node.hasFlag = false
			}
		}
		if len(flowToUpload.nodes) > 0 {
			flowToUpload.start = flowToUpload.nodes[0].time
			flowToUpload.end = flowToUpload.nodes[len(flowToUpload.nodes)-1].time

			log.Debug(flowToUpload.String())

			/*UPLOAD FLOWT TO MONGO HERE*/
			// flowToUpload.uploadToMongo()
		} else {
			/*dont upload to mongo empty flows*/
			log.Warning("No nodes found for flow [" + flowToUpload.flowID + "]" +
				": Connection was reset right after the 3-way-hand-shake")
		}
	}
}

// ReassemblyComplete is called when the TCP assembler
// believes a stream has finished.
func (s *uniStream) ReassemblyComplete() {
	log.Debugf("Reassembly of stream %v:%v complete ", //- start:%v end:%v bytes:%v",
		s.bothStrims.key.net, s.bothStrims.key.transport) // s.start, s.end, s.bytes)

	s.done = true
	s.bothStrims.maybeFinish()
}
