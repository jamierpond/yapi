.PHONY: build install clean

GOBIN ?= $(shell go env GOPATH)/bin
TARGET = imgcat

build:
	@echo "Building $(TARGET)..."
	@go build -o $(TARGET) ./cmd

install:
	@echo "Installing $(TARGET) to $(GOBIN)..."
	@go install ./cmd

clean:
	@echo "Cleaning..."
	@rm -f $(TARGET)
