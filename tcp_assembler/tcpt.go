package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
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

type dataFlowt struct {
	size int64
	// some redundancy for faster processing
	data string // printable representation of the data
	hex  []byte // hex representation of the data
}

type flowt struct {
	id               string
	srcIP, dstIP     string
	srcPort, dstPort uint16
	time             int64 // as is returned by time.Now().UnixNano()
	hasFlag          bool  // regex find for flag{...} pattern
	favourite        bool  /* defaults to false, can only be
	changed from the front-end*/
	hasSYN, hasFIN bool
	dataFlow       dataFlowt // custom type
}

// tcpStreamFactory implements tcpassembly.StreamFactory
type tcpStreamFactory struct{}

// tcpStream will handle tcp packets
type tcpStream struct {
	net, transport   gopacket.Flow
	bytes            int64
	payload          []byte
	start            time.Time
	sawStart, sawEnd bool
}

func (factory *tcpStreamFactory) New(net, tport gopacket.Flow) tcpassembly.Stream {
	log.Tracef("new stream %v:%v started", net, tport)
	t := &tcpStream{
		net:       net,
		transport: tport,
		start:     time.Now(),
	}
	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return t
}

// Reassembled is called whenever new packet data is available for reading.
// Reassembly objects contain stream data IN ORDER.
func (s *tcpStream) Reassembled(reassemblies []tcpassembly.Reassembly) {
	for _, reassembly := range reassemblies {
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
		dataFlow := &dataFlowt{
			size: s.bytes,
			hex:  s.payload,
		}
		temp := make([]byte, len(s.payload))
		for i, octet := range s.payload {
			//if character is printable, add it; else add a "."
			if IsASCIIPrintable(rune(octet)) {
				temp[i] = byte(octet)
			} else {
				temp[i] = 0x2e
			}
		}
		dataFlow.data = string(temp)

		flowToUpload := &flowt{
			srcIP:     s.net.Src().String(),
			dstIP:     s.net.Dst().String(),
			srcPort:   bytesToUint16(s.transport.Src().Raw()),
			dstPort:   bytesToUint16(s.transport.Dst().Raw()),
			hasFlag:   false,
			favourite: false,
			hasSYN:    s.sawStart,
			hasFIN:    s.sawEnd,
			time:      s.start.UnixNano(),
			dataFlow:  *dataFlow,
		}

		/* Concatenate all unique info of the flow in order
		to make a unique MD5 which will be the flow's id */
		t := append(s.net.Src().Raw(), s.net.Dst().Raw()...)
		t = append(t, s.transport.Src().Raw()...)
		t = append(t, s.transport.Dst().Raw()...)
		// create a temporary buffer to hold the time converted from int64 to []byte
		bytetime := new(bytes.Buffer)
		binary.Write(bytetime, binary.LittleEndian, flowToUpload.time)
		t = append(t, bytetime.Bytes()...)
		t = append(t, flowToUpload.dataFlow.hex...)
		md5sum := md5.Sum(t)

		flowToUpload.id = hex.EncodeToString(md5sum[:])

		log.Info("id:", flowToUpload.id)
		log.Info("srcIP:", flowToUpload.srcIP)
		log.Info("dstIP:", flowToUpload.dstIP)
		log.Info("srcPort:", flowToUpload.srcPort)
		log.Info("dstPort:", flowToUpload.dstPort)
		log.Info("hasFlag:", flowToUpload.hasFlag)
		log.Info("favourite:", flowToUpload.favourite)
		log.Info("hasSYN, hasFIN: ", flowToUpload.hasSYN, ", ", flowToUpload.hasFIN)
		log.Info("time:", time.Unix(0, flowToUpload.time))
		log.Info("dataFlow.size:", flowToUpload.dataFlow.size)
		log.Info("dataFlow.data:", flowToUpload.dataFlow.data)
		log.Info("-------------------------------\n")

		/*UPLOAD FLOWT TO MONGO HERE*/
	}
}
