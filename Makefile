GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

BIN_DIR=bin
SRC_DIR=tcp_assembler
TCP_A=$(BIN_DIR)/tcp_assembler

.PHONY: clean examples

all:
	@cd $(SRC_DIR)
	$(GOBUILD) -o $(TCP_A) $(SRC_DIR)/main.go $(SRC_DIR)/tcpt.go $(SRC_DIR)/db.go

tfp:
	$(GOBUILD) -o ../bin/tcpflow_parser tcpflow_parser/tcpflow_parser.go tcpflow_parser/db.go 


clean:
	rm -rf $(BIN_DIR)
