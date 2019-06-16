# The Idea

The project will be divided into four components:

* **tcpdump** - We will run `tcpdump` on the vulnbox as follows:
```
# tcpdump -i [device] -G 60 -z "./send_to_remote_assembler_and_archive.py" -w dump_%Y-%m-%d_%H:%M:%S.pcap tcp port 8080 or 443 or 80
```
* **tcp_assembler** - Written in Go. This is a service that periodically checks for new files named `dump_%Y-%m-%d_%H:%M:%S.pcap`, processes them, then archives them.
* **The Database** - This will be where all the data from the TCP streams will be archived. Ideally, we will be using a NoSQL DB (mongodb), that can be installed on any machine, but, by default, it will be assumed to be running on the same machine as the Packet Sniffer, on port 27017. I don't think much coding will be required for this component, it should *just work* ™.
* **The Front-end** - This will be the interface through which we will access the data stored in the Database. It can be realized either on python flask, any node-js server, or Nginx. This component will need to have the following features: **real-time traffic updates** (done through ajax requests), **filter by "presence of a flag"** in a TCP stream, **ease of use** in order to make other team's exploit reusable as fast as possible (see how [Flower](https://github.com/secgroup/flower) does it).

Suggestions are very much appreciated, on our Telegram group.
---
# The Minimum Viable Product (MVP)

### tcp_assembler ✓

### Database ✓

### Webserver ✓
---
# Next steps

### tcp_assembler
* ... 

### Webserver
* screen to show current iptables configuration
* option to manually mark bad packet contents with `iptables`, and automatically update iptables' configuration on the vulnbox
* ...


### Other
* make a new packet sniffer to provide heuristic analysis to mark bad packets
* ...
---
# Flow Structure
`flowt` data structure, in Go:
```go
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
```

This structure will be uploaded to mongodb as follows:
```json
"flows": [
    {
        "_id": "6f7197b90c28d1cafd730b82d0ca8f27", //generated by mongo
        "time": NumberLong(53492),
        "duration": // flowt.end - flowt.start,
        "srcIP": "192.168.1.133",
        "srcPort": 1234 ,
        "dstIP": "127.0.0.1",
        "dstPort": 1234,
        "hasFlag": false,
        "trafficSize": 14135, //measured in bytes
        "favorite": false,
        "seenSYN": false,
        "seenFIN": true,
        "nodes": [
            {
                "fromSrc": true,
                "time": NumberLong(53493)
                "printableData": "...",
                "blob": BinData(0,"FwMDACLFPqgef8h024g08hg4g208y="),
                "hasFlag": false
            },
            ...
        ]
    },
    ...
]
```
# Mongodb Usage

Start the db with:
```pseudocode
mongod --dbpath /path/to/where_you_want_your_db_to_be
```

Connect to the db process with:
```pseudocode
mongo
```
Followed by:
```pseudocode
use my_db
```

## Use these commands to perform various tests:

Declare these variables first:
```pseudocode
r = db.getCollection("connections")
c = db.getCollection("flows")
```
See connections:
```pseudocode
r.find().pretty()
```
See flows:
```pseudocode
c.find().pretty()
```
Remove all data from the db:
```pseudocode
r.deleteMany({})
c.deleteMany({})
```
---
# Build Requirements

* `go` version >=1.12
* Arch Linux dependencies: `libpcap`
* Ubuntu dependencies: `libpcap-dev`

Before building for the first time, you will need to run the following commands:

```pseudocode
$ go get github.com/google/gopacket
$ go get github.com/sirupsen/logrus
$ go get go.mongodb.org/mongo-driver/mongo
```
---
# Install Instructions

### Download
```pseudocode
$ git clone https://gitlab.com/cc19-sapienza/timon.git
```
### Build
```pseudocode
$ make
```
### Run
```pseudocode
$ sudo tcpdump -i enp0s31f6 -w - "tcp port 8080 or 443 or 80" | ./bin/tcp_assembler -r - -debug
```
or
```pseudocode
$ ./bin/tcp_assembler -r pcaps/dump_2019-06-14_15\:45\:30.pcap -debug
```
---
# Production Run

1. Open 4 terminals
2. cd into bin/ , then `mkdir pcaps archive`
3. on one terminal, cd into bin/pcaps , the run `sudo tcpdump -i enp0s31f6 -G 60 -w dump_%Y-%m-%d_%H:%M:%S.pcap "tcp port 8080 or 443 or 80"` as root, changing interface and filter if necessary
4. on the second terminal run `mongod --dbpath /path/to/where/you/want/your/db`
5. on the third terminal, cd into bin/, then run `./tcp_assembler -nodebug`
6. on the last terminal, cd into webserver and run `FLASK_APP=webserver.py flask run`

DONE