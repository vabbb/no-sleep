package main

import (
	"fmt"
	"log"
	"time"
	"unicode"

	"github.com/golang/glog"
	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
)

// IsASCIIPrintable will return true if char is printable
// else it will return false
func IsASCIIPrintable(r rune) bool {
	if r > unicode.MaxASCII || !unicode.IsPrint(r) {
		return false
	}
	return true
}

type dataFlowt struct {
	from string
	// some redundancy for faster processing
	data string // printable representation of the data
	hex  []byte // hex representation of the data
	time int64  // as is returned by time.Now().UnixNano()
}

type flowt struct {
	srcIP, dstIP     string
	srcPort, dstPort uint16
	time             int64 // as is returned by time.Now().UnixNano()
	lastSeen         int64 // also in nanoseconds
	hasFlag          bool  // regex find for flag{...} pattern
	favourite        bool  // defaults to false, can only be
	// changed from the front-end
	dataFlow []dataFlowt // custom type
}

// tcpStreamFactory implements tcpassembly.StreamFactory
type tcpStreamFactory struct{}

// tcpStream will handle tcp packets
type tcpStream struct {
	net, transport   gopacket.Flow
	bytes            int64
	payload          []byte
	start, end       time.Time
	sawStart, sawEnd bool
}

func (factory *tcpStreamFactory) New(net, tport gopacket.Flow) tcpassembly.Stream {
	log.Printf("new stream %v:%v started", net, tport)
	t := &tcpStream{
		net:       net,
		transport: tport,
		start:     time.Now(),
	}
	t.end = t.start
	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return t
}

// Reassembled is called whenever new packet data is available for reading.
// Reassembly objects contain stream data IN ORDER.
func (s *tcpStream) Reassembled(reassemblies []tcpassembly.Reassembly) {
	for _, reassembly := range reassemblies {
		if !reassembly.Seen.Before(s.end) {
			s.end = reassembly.Seen
		}
		s.payload = reassembly.Bytes
		s.bytes += int64(len(reassembly.Bytes))

		s.sawStart = s.sawStart || reassembly.Start
		s.sawEnd = s.sawEnd || reassembly.End
	}
}

// ReassemblyComplete is called when the TCP assembler believes a stream has
// finished.
func (s *tcpStream) ReassemblyComplete() {
	diffSecs := float64(s.end.Sub(s.start)) / float64(time.Second)
	glog.V(2).Infof("Reassembly of stream %v:%v complete - start:%v end:%v bytes:%v bps:%v",
		s.net, s.transport, s.start, s.end, s.bytes,
		float64(s.bytes)/diffSecs)
	fmt.Println("Payload:")
	for i, octet := range s.payload {
		//if character is printable, print it; else print a "."
		if IsASCIIPrintable(rune(octet)) {
			fmt.Print(string(octet))
		} else {
			fmt.Print(".")
		}
		//every 80 characters printed, print a newline
		if (i+1)%80 == 0 {
			fmt.Println()
		}
	}
	fmt.Print("\n-------------------------------\n")
}
