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

.PHONY: build test clean