.PHONY: build run run-print-analytics test fmt fmt-check clean install docker web web-run bump-patch bump-minor bump-major release build-all

NAME := yapi
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
# PostHog keys - set via environment or CI/CD secrets
# These are write-only project keys, but kept out of git history for cleanliness
POSTHOG_API_KEY ?=
POSTHOG_API_HOST ?= https://us.i.posthog.com
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) \
	-X yapi.run/cli/internal/telemetry.PosthogAPIKey=$(POSTHOG_API_KEY) \
	-X yapi.run/cli/internal/telemetry.PosthogAPIHost=$(POSTHOG_API_HOST)

install: build
	@echo "Installing yapi to $$(go env GOPATH)/bin..."
	@cp ./bin/yapi $$(go env GOPATH)/bin/yapi
	@codesign --sign - --force $$(go env GOPATH)/bin/yapi 2>/dev/null || true
	@echo "Done! Ensure $$(go env GOPATH)/bin is in your PATH."

lint:
	@echo "Running deadcode analysis..."
	@deadcode ./...

build:
	@echo "Building yapi CLI..."
	@go build -ldflags "$(LDFLAGS)" -o ./bin/yapi ./cmd/yapi
	@codesign --sign - --force ./bin/yapi 2>/dev/null || true

run:
	@echo "Running yapi CLI..."
	@go run ./cmd/yapi

run-print-analytics: build
	@echo "Running yapi CLI with analytics printing..."
	@YAPI_PRINT_ANALYTICS=1 ./bin/yapi $(RUN_ARGS)

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

bump-patch:
	@./scripts/bump.sh patch

bump-minor:
	@./scripts/bump.sh minor

bump-major:
	@./scripts/bump.sh major

release:
	@TAG=$$(git describe --tags --abbrev=0); \
	echo "Pushing $$TAG to origin..."; \
	git push origin "$$TAG"
