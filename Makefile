GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

BIN_DIR=./bin
SRC_DIR=./tcp_assembler
TCP_A=$(BIN_DIR)/tcp_assembler
TARGETS=$(SRC_DIR)/*.go

.PHONY: clean examples

all:
	$(GOBUILD) -o $(TCP_A) $(TARGETS)

clean:
	rm -rf $(BIN_DIR)