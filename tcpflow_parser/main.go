package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type nodet struct {
	connID           string
	nodeID           string
	srcIP, dstIP     string
	srcPort, dstPort uint16
	hasFlag          bool
	time             int64
	data             string
	hex              []byte
}

//this only works when piped with tcpflow
func main() {

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
