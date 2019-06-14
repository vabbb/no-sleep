GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

BIN_DIR=bin
SRC_DIR=tcp_assembler
SOURCES=$(wildcard $(SRC_DIR)/*.go)
TCP_A=$(BIN_DIR)/tcp_assembler

.PHONY: clean examples

all:
	$(GOBUILD) -o $(TCP_A) $(SOURCES)

bidi:
	$(GOBUILD) -o $(BIN_DIR)/bidi examples/bidi.go

clean:
	rm -rf $(BIN_DIR)
