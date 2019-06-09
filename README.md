# The Idea

The project will be divided into four components:

* **tcpdump** - We will run `tcpdump` on the vulnbox as follows:

      # tcpdump -i [device] -G 60 -z "./send_to_remote_assembler_and_archive.py" -w dump_%Y-%m-%d_%H:%M:%S.pcap tcp port 8080 or 443 or 80

* **tcp_assembler** - Written in Go. This is a service that periodically checks for new files named `dump_%Y-%m-%d_%H:%M:%S.pcap`, processes them, then archives them.
* **The Database** - This will be where all the data from the TCP streams will be archived. Ideally, we will be using a NoSQL DB (mongodb), that can be installed on any machine, but, by default, it will be assumed to be running on the same machine as the Packet Sniffer, on port 27017. I don't think much coding will be required for this component, it should *just work* ™.
* **The Front-end** - This will be the interface through which we will access the data stored in the Database. It can be realized either on python flask, any node-js server, or Nginx. This component will need to have the following features: **real-time traffic updates** (done through ajax requests), **filter by "presence of a flag"** in a TCP stream, **ease of use** in order to make other team's exploit reusable as fast as possible (see how [Flower](https://github.com/secgroup/flower) does it).

Suggestions are very much appreciated, on our Telegram group.

# The Minimum Viable Product (MVP)

### tcp_assembler
* ✓ ~~able to parse data from pcap files~~
* ✓ ~~able to look for pcap files to process (with the format described in the above paragraph)... for now, we are fine with it just sorting pcap files by date, then processing the oldest one, then moving it to an `archive/` folder and start over again.~~
* ✓ ~~able to assemble data belonging to a single tcp flow~~
* able to recognize when a flow is completed
* able to assign the completed flow to the `flowt` struct
* able to push a finished tcp flow to mongodb, using the stuct defined later, plus a unique id (which can be the hash of the tcp flow identifiers and its start time)

### Database
* having a working instance of `mongodb`

### Webserver
* bare-bones webserver able to query the mongodb and display the data that was pushed to it

# Next steps

### tcp_assembler
* parsing flow contents to check for presence of flags
* ... 

### Webserver
* screen to show current iptables configuration
* option to manually mark bad packet contents with `iptables`, and automatically update iptables' configuration on the vulnbox
* ...


### Other
* make a new packet sniffer to provide heuristic analysis to mark bad packets
* ...

# Flow Structure
`flowt` data structure, in Go:

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


With `dataFlowt` being like this:

    type dataFlowt struct {
        from string
        // some redundancy for faster processing
        data string // printable representation of the data
        hex  []byte // hex representation of the data
        time int64  // as is returned by time.Now().UnixNano()
    }


# Build Requirements

* `go` version >=1.12
* Arch Linux dependencies: `libpcap`
* Ubuntu dependencies: `libpcap-dev`

Before building for the first time, you will need to run the following commands:

    $ go get github.com/google/gopacket
    $ go get github.com/sirupsen/logrus

# Install Instructions

### Download
    $ git clone https://gitlab.com/cc19-sapienza/timon.git

### Build
    $ make

### Run
    $ ./bin/tcp_assembler [-d pcaps' directory]
