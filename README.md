# The Idea

The project will be divided into three components:

* **tcpdump** - We will use tcpdump as follows: `# tcpdump -G 120 -w dump-%H:%M:%S.pcap -z "./script.py" tcp port 8080 or 443 or 80 &` . This way, it will create a new pcap file every 120 seconds, executing a certain script onas `script.py file` every time it finishes writing to a pcap file.
* **The TCP Flow Assembler** - This will assemble the TCP stream and put it into a custom struct, called "flow". This data will be uploaded to the database as soon as it's processed.
* **The Database** - This will be where all the data from the TCP streams will be archived. Ideally, we will be using a NoSQL DB (mongodb), that can be installed on any machine, but, by default, it will be assumed to be running on the same machine as the Packet Sniffer, on port 27017. I don't think much coding will be required for this component, it should *just work* ™.
* **The Front-end** - This will be the interface through which we will access the data stored in the Database. It can be realized either on python flask, any node-js server, or Nginx. This component will need to have the following features: **real-time traffic updates** (done through ajax requests), **filter by "presence of a flag"** in a TCP stream, **ease of use** in order to make other team's exploit reusable as fast as possible (see how [Flower](https://github.com/secgroup/flower) does it).

Suggestions are very much appreciated, on our Telegram group.

# Build Requirements

* `go` version >=1.12
* Arch Linux dependencies: `libpcap`
* Ubuntu dependencies: `libpcap-dev`

# Install Instructions

### Download
* `git clone https://gitlab.com/cc19-sapienza/timon.git`

### Build
* `make`

### Run
* `sudo timon -i [interface]`

(The `timon` executable is in the `bin/` folder)
