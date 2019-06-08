GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

BIN_DIR=./bin
SRC_DIR=./flow_parser
FLOW_P=$(BIN_DIR)/flow_parser
TARGETS=$(SRC_DIR)/*.go

.PHONY: clean examples

all:
	$(GOBUILD) -o $(FLOW_P) $(TARGETS)

clean:
	rm -rf $(BIN_DIR)