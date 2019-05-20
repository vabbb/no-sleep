# Requirements

* `go` version >=1.11
* Arch Linux dependencies: `libpcap`
* Ubuntu dependencies: `libpcap-dev`

# Install Instructions

### Download
`git clone https://gitlab.com/cc19-sapienza/timon.git`

### Build
`go build -o bin/timon ./...`

### Run
`sudo timon [interface]`

(The `timon` executable is in the `bin/` folder)

In order to run timon, you need to choose an interface to monitor.

You can list all of your interfaces with the command `ip addr`