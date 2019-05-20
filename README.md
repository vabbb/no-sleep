# Requirements

* go version >=1.11

# Install Instructions

## Clone
`https://gitlab.com/cc19-sapienza/timon.git`

## Build
`go build -o bin/timon ./...`

## Run
`sudo bin/timon [interface]`

In order to run timon, you need to choose an interface to monitor.

You can list all of your interfaces with the command `ip addr`