# OUTLINE

### The pcap Importer Thread
This is a thread that is meant to be checking the pcaps folder for new files to be imported. It imports a pcap **iff** there are 2 pcaps in the folder, one that is done being written on (which will be imported), and one that just stared being written.

----

### The TCP Flow Assembler
We define TCP Flow to be a struct, containing the following data:

    type flow struct {
        filename    string      // name of the pcap file from which this
                                // flow started
        src_ip      string      // IP which started the TCP handshake
        dst_ip      string      // IP on the receiving end of the TCP handshake
        src_port    uint16
        dst_port    uint16
        start_time  int64       // as is returned by time.Now().UnixNano()
                                // measured in nanoseconds
        duration    int64       // this too is measured in nanoseconds
        has_flag    bool        // regex find for flag{...} pattern
        favourite   bool        // defaults to false, can only be
                                // changed from the front-end
        data_flow   []data_flowt    // custom type
    }

With data_flowt being like this:

    type data_flowt struct {
        from    bool        // true if from src_ip:src_port
                            // false if from dst_ip:dst_port
        data    string      // printable representation of the data
        hex     []byte      // hex representation of the data
        time    int64       // as is returned by time.Now().UnixNano()
    }

----

### Multithreading

For every new flow, identified by a unique quintuple of "src_ip, dst_ip, src_port, dst_port, start_time", a new thread is started, which will be given all data belonging to this particular flow.

It is this thread's duty to **reassemble** and **process** all packets belonging to this flow, and store them into mongodb as quickly as possible.

The thread will die after 5 minutes (and will **forcefully** close the connection), or after the TCP connection is closed.