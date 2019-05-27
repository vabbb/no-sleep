GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

BIN_DIR=./bin
SRC_DIR=./src
TIMON=$(BIN_DIR)/timon
TARGETS=$(SRC_DIR)/*.go

.PHONY: clean examples

all:
	$(GOBUILD) -o $(TIMON) $(TARGETS)

examples:
	$(GOBUILD) -o $(BIN_DIR)/httpassembly $(SRC_DIR)/httpassembly.go
	$(GOBUILD) -o $(BIN_DIR)/statsassembly $(SRC_DIR)/statsassembly.go
	$(GOBUILD) -o $(BIN_DIR)/tcpassembly $(SRC_DIR)/tcpassembly.go

clean:
	rm -rf $(BIN_DIR)