# yapi Development Guidelines

## Project Overview

yapi is a CLI-first, git-friendly API client for HTTP, gRPC, GraphQL, and TCP.

## Core Principles

1. **CLI-First**: All features via terminal, scriptable and pipeable
2. **Git-Friendly**: YAML configs, no binary formats
3. **Protocol Agnostic**: HTTP, gRPC, GraphQL, TCP as equals
4. **Simplicity**: YAGNI, minimal defaults

## Project Structure

```text
cmd/yapi/           # CLI entry point
internal/
├── executor/       # Protocol executors (http, grpc, graphql, tcp)
├── config/         # YAML parsing, environments
├── tui/            # Interactive mode
└── lsp/            # Language server
examples/           # Sample .yapi.yml files
tests/              # Integration tests
```

## Commands

```bash
make build          # Build binary
make test           # Run tests
make lint           # Run linter
make install        # Install locally
yapi run file.yapi.yml      # Execute request
yapi validate file.yapi.yml # Validate schema
```

## Code Style

- Table-driven tests
- Error messages must be actionable
- Keep packages small and focused
- Prefer explicit over implicit
