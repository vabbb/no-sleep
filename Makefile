GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

BIN_DIR=bin
SRC_DIR=tcp_assembler
TCP_A=$(BIN_DIR)/tcp_assembler

.PHONY: clean examples

all:
	@cd $(SRC_DIR)
	$(GOBUILD) -o $(TCP_A) ./...

tfp:
	@cd tcpflow_parser
	$(GOBUILD) -o ../bin/tcpflow_parser main.go db.go


clean:
	rm -rf $(BIN_DIR)
