# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]

## Summary

[Primary requirement + technical approach]

## Technical Context

**Language**: Go 1.21+
**Key Packages**: `internal/executor`, `internal/config`, `internal/tui`
**Testing**: `go test`, table-driven tests
**Build**: `make build && make test && make lint`

## Constitution Check

*Must pass before implementation.*

| Principle | Status | Notes |
|-----------|--------|-------|
| CLI-First | [ ] | Feature accessible via `yapi` command? |
| Git-Friendly | [ ] | Config in YAML? No binary formats? |
| Protocol Agnostic | [ ] | Works across HTTP/gRPC/GraphQL/TCP? |
| Simplicity | [ ] | Minimal implementation? No over-engineering? |
| Dogfooding | [ ] | Can webapp use this via yapi? |

## Affected Areas

```text
cmd/yapi/           # CLI entry points
internal/
├── executor/       # HTTP, gRPC, TCP, GraphQL execution
├── config/         # YAML parsing, environment handling
├── tui/            # Interactive mode (optional)
└── [new-pkg]/      # New package if needed
```

## Implementation Approach

[Describe the approach in 3-5 bullet points]

## Complexity Justification

> Fill only if Constitution Check has concerns

| Concern | Why Needed | Simpler Alternative Rejected |
|---------|------------|------------------------------|
