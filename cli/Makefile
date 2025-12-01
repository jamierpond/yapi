build:
	@echo "Building yapi CLI..."
	@go build -o ./cmd/yapi/yapi ./cmd/yapi/main.go

run:
	@echo "Running yapi CLI..."
	@go run ./cmd/yapi

test:
	@echo "Running all tests..."
	@go test ./...

clean:
	@echo "Cleaning up..."
	@rm -f ./cmd/yapi/yapi
	@go clean

install: build
	@echo "Installing yapi to $(shell go env GOPATH)/bin..."
	@cp ./cmd/yapi/yapi $(shell go env GOPATH)/bin/yapi
	@echo "Done! Ensure $(shell go env GOPATH)/bin is in your PATH."

.PHONY: build run test clean install