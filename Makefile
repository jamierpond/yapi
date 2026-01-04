.PHONY: build run run-print-analytics test fuzz fmt fmt-check clean install docker web web-run bump-patch bump-minor bump-major release build-all lint lint-install install-lint lint-quick lint-full gen-docs gh-action fuzz-cover local-release

NAME := yapi
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)


install: build
	@echo "Installing yapi to $$(go env GOPATH)/bin..."
	@cp ./cli/bin/yapi $$(go env GOPATH)/bin/yapi
	@codesign --sign - --force $$(go env GOPATH)/bin/yapi 2>/dev/null || true
	@echo "Done! Ensure $$(go env GOPATH)/bin is in your PATH."

kore:
	yapi import ./postman-examples/kore.ai/collection.json -e ./postman-examples/kore.ai/env.json --output foo


fuzz-cover:
	@cd cli && go test ./... -run=Fuzz -coverprofile=fuzz.cov
	@cd cli && go tool cover -func=fuzz.cov

# Install linting tools
lint-install:
	@echo "Installing linting tools..."
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
	go install golang.org/x/vuln/cmd/govulncheck@latest

# Quick lint (go vet + fmt check)
lint-quick:
	@echo "Running go vet..."
	@cd cli && go vet ./...
	@echo "Checking gofmt..."
	@test -z "$$(gofmt -s -l cli/cmd cli/internal)" || (echo "Files need formatting:"; gofmt -s -l cli/cmd cli/internal; exit 1)

# Standard lint (golangci-lint with all enabled linters)
lint:
	@echo "Running golangci-lint..."
	@cd cli && golangci-lint run ./...

# Full lint (includes vulnerability check)
lint-full: lint
	@echo "Running govulncheck..."
	@cd cli && govulncheck ./...

build:
	@echo "Building yapi CLI..."
	@cd cli && go build -ldflags "$(LDFLAGS)" -o ./bin/yapi ./cmd/yapi
	@codesign --sign - --force ./cli/bin/yapi 2>/dev/null || true

run:
	@echo "Running yapi CLI..."
	@cd cli && go run ./cmd/yapi

run-print-analytics: build
	@echo "Running yapi CLI with analytics printing..."
	@YAPI_PRINT_ANALYTICS=1 ./cli/bin/yapi $(RUN_ARGS)

test:
	@echo "Running all tests..."
	@cd cli && go test -cover -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage summary:"
	@cd cli && go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'
	@rm -f cli/coverage.out

fuzz:
	@cd cli && go run ./scripts/fuzz.go

fmt:
	@echo "Formatting code..."
	@gofmt -w ./cli

fmt-check:
	@echo "Checking formatting..."
	@test -z "$$(gofmt -l cli/cmd cli/internal)" || (echo "Files not formatted:"; gofmt -l cli/cmd cli/internal; exit 1)

clean:
	@echo "Cleaning up..."
	@rm -f ./cli/bin/yapi
	@cd cli && go clean


web:
	docker build . -t ${NAME}:latest -f Dockerfile.webapp


web-run:
	-docker stop yapi
	-docker rm yapi
	docker run --name yapi -p 3000:3000 ${NAME}:latest

bump-patch:
	@./cli/scripts/bump.sh patch

bump-minor:
	@./cli/scripts/bump.sh minor

bump-major:
	@./cli/scripts/bump.sh major

release:
	@BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$BRANCH" != "main" ] && [ "$$BRANCH" != "develop" ] && [ "$$BRANCH" != "next" ]; then \
		echo "Error: Releases can only be made from 'main', 'develop', or 'next' branches"; \
		echo "Current branch: $$BRANCH"; \
		exit 1; \
	fi
	@echo "Pushing commits and tags to origin..."
	@git push origin HEAD
	@TAG=$$(git describe --tags --abbrev=0); \
	echo "Pushing tag $$TAG..."; \
	git push origin "$$TAG"
	@echo "Release complete!"

gen-docs:
	@echo "Generating CLI documentation..."
	@cd cli && go run scripts/gendocs.go


gh-action:
	@echo "Running tests for GitHub Actions..."
	act -W .github/workflows/web-tests.yml \
		--container-architecture linux/amd64

local-release:
	goreleaser release --snapshot --clean --skip=publish
