# The Idea

The project will be divided into three components:

* **The Packet Sniffer** - This will capture TCP packets on one interface, it will be able to monitor multiple ports (chosen by the user), and reassemble the TCP stream. This data will be uploaded to the database.
* **The Database** - This will be where all the data from the TCP streams will be archived. Ideally, we will be using a NoSQL DB (mongodb), that can be installed on any machine, but, by default, it will be assumed to be running on the same machine as the Packet Sniffer, on port 27017. I don't think much coding will be required for this component.
* **The Front-end** - This will be the interface through which we will access the data stored in the Database. It can be realized either on python flask, any node-js server, or Nginx. This component will need to have the following features: **real-time traffic updates** (done through ajax requests), **filter by "presence of a flag"** in a TCP stream, **ease of use** in order to make other team's exploit reusable as fast as possible (see how [Flower](https://github.com/secgroup/flower) does it).

Suggestions are very much appreciated, on our Telegram group.

# Requirements

* `go` version >=1.12
* Arch Linux dependencies: `libpcap`
* Ubuntu dependencies: `libpcap-dev`

# Install Instructions

### Download
`git clone https://gitlab.com/cc19-sapienza/timon.git`

### Build
`go build -o bin/timon timon.go`

### Run
`sudo timon [interface]`

(The `timon` executable is in the `bin/` folder)

In order to run timon, you need to choose an interface to monitor.

You can list all of your interfaces with the command `ip addr`
