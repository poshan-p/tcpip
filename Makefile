# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Name of your executable
BINARY_NAME=tcpip

all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/main/main.go

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/main/main.go
	./$(BINARY_NAME)

.PHONY: all build clean run

