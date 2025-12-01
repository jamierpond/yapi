.PHONY: build run test fmt fmt-check clean install docker web web-run

NAME := yapi

install: build
	@echo "Installing yapi to $$(go env GOPATH)/bin..."
	@cp ./bin/yapi $$(go env GOPATH)/bin/yapi
	@codesign --sign - --force $$(go env GOPATH)/bin/yapi 2>/dev/null || true
	@echo "Done! Ensure $$(go env GOPATH)/bin is in your PATH."

build:
	@echo "Building yapi CLI..."
	@go build -o ./bin/yapi ./cmd/yapi
	@codesign --sign - --force ./bin/yapi 2>/dev/null || true

run:
	@echo "Running yapi CLI..."
	@go run ./cmd/yapi

test:
	@echo "Running all tests..."
	@go test ./...

fmt:
	@echo "Formatting code..."
	@gofmt -w .

fmt-check:
	@echo "Checking formatting..."
	@test -z "$$(gofmt -l cmd internal)" || (echo "Files not formatted:"; gofmt -l cmd internal; exit 1)

clean:
	@echo "Cleaning up..."
	@rm -f ./bin/yapi
	@go clean


web:
	docker build . -t ${NAME}:latest -f Dockerfile.webapp

web-run:
	-docker stop yapi
	-docker rm yapi
	docker run --name yapi -p 3000:3000 ${NAME}:latest
