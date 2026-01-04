# yapi Development Guidelines

## Project Overview

yapi is a CLI-first, git-friendly API client for HTTP, gRPC, GraphQL, and TCP.

## Core Principles

1. **CLI-First**: All features via terminal, scriptable and pipeable
2. **Git-Friendly**: YAML configs, no binary formats
3. **Protocol Agnostic**: HTTP, gRPC, GraphQL, TCP as equals
4. **Simplicity**: YAGNI, minimal defaults
5. **Aggressive Code Removal**: Delete unused code immediately. No deprecated functions, no backwards-compat shims, no "just in case" abstractions. Can't have bugs in code you don't have.
6. **Single Maintainer Mindset**: This is maintained by one person. Every line of code is a liability. Less code = fewer bugs = easier maintenance.

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

## Code Hygiene

- **No deprecated functions**: When you change an API, update all callers and delete the old function. Don't keep deprecated wrappers around.
- **No backwards-compatibility shims**: If something is unused, delete it completely. No `_unused` renames, no re-exports, no `// removed` comments.
- **One code path**: If two pieces of code do the same thing, consolidate them. Duplicate logic is duplicate bugs.
- **Delete tests for deleted code**: When you delete a function, delete its tests too.
- **Fewer lines is better**: Given two correct solutions, prefer the one with less code.
