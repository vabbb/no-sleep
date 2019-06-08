# The Idea

The project will be divided into three components:

* **tcpdump & tcpflow** - We will use tcpdump and tcpflow as follows:
    
        $ ssh -t root@vulnbox "tcpdump -i [device] -w - tcp port 8080 or 443 or 80" | tcpflow -r - -c -D | flow_parser 

* **flow_parser** - Written in Go. A simple program that will parse the tcpflow output with regexes (also looking for the presence of flags) and upload a flow structure to the database as soon as it's processed.
* **The Database** - This will be where all the data from the TCP streams will be archived. Ideally, we will be using a NoSQL DB (mongodb), that can be installed on any machine, but, by default, it will be assumed to be running on the same machine as the Packet Sniffer, on port 27017. I don't think much coding will be required for this component, it should *just work* ™.
* **The Front-end** - This will be the interface through which we will access the data stored in the Database. It can be realized either on python flask, any node-js server, or Nginx. This component will need to have the following features: **real-time traffic updates** (done through ajax requests), **filter by "presence of a flag"** in a TCP stream, **ease of use** in order to make other team's exploit reusable as fast as possible (see how [Flower](https://github.com/secgroup/flower) does it).

Suggestions are very much appreciated, on our Telegram group.

# The Minimum Viable Product (MVP)

* `flow_parser` able to parse data from tcpflow and push it to mongodb
* having a working instance of mongodb
* bare-bones web-server able to query the mongodb and display the data that was pushed to it

### Pushing flows to database:
* If a flow with the same identifiers (ip addresses and port addresses) already exists in the database, check for its last_seen field. If it was last_seen within the past 5 minutes, add the current flow to that one. Otherwise, create a new flow.

# Next steps

### flow_parser
* parsing flow contents to check for presence of flags
* ... 

### Web-Server
* screen to show current iptables configuration
* option to manually mark bad packet contents with iptables
* ...
* heuristic analysis to mark bad packets
* ...

# Flow Structure
Flow data structure, in Go:

    type flow struct {
        src_ip      string
        dst_ip      string
        src_port    uint16
        dst_port    uint16
        time        int64       // as is returned by time.Now().UnixNano()
                                // measured in nanoseconds
        last_seen   int64       // in nanoseconds
        has_flag    bool        // regex find for flag{...} pattern
        favourite   bool        // defaults to false, can only be
                                // changed from the front-end
        data_flow   []data_flowt    // custom type
    }

With data_flowt being like this:

    type data_flowt struct {
        from    string
        // some redundancy for faster processing
        data    string      // printable representation of the data
        hex     []byte      // hex representation of the data
        time    int64       // as is returned by time.Now().UnixNano()
    }


# Build Requirements

* `go` version >=1.12

# Install Instructions

### Download
* `git clone https://gitlab.com/cc19-sapienza/timon.git`
