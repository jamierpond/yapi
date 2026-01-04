# Quickstart: API Consolidation Refactor

**Date**: 2026-01-03
**Feature**: 002-api-consolidation-refactor

## Prerequisites

- Go 1.21+
- Make

## Build & Test

```bash
# Build
make build

# Run tests
make test

# Run linter
make lint
```

## Implementation Order

Execute phases in order. Run `make test && make lint` after each phase.

### Phase 1: AST Extraction

1. Create `internal/validation/ast.go`
2. Move `findVarPositionInYAML`, `findNodeInMapping`, `findKeyNodeInMapping` from langserver
3. Update langserver to use validation.FindVarPositionInYAML
4. Update analyzer to use AST helpers for line numbers

**Verify**: `make test && make lint`

### Phase 2: Analyzer Consolidation

1. Add `AnalyzeOptions` struct to analyzer.go
2. Add `Analyze(text string, opts AnalyzeOptions)` function
3. Refactor `analyzeParsed` to accept opts
4. Add deprecation wrappers for old functions

**Verify**: `make test && make lint`

### Phase 3: Executor Simplification

1. Add `GetTransport(transport, client)` function
2. Move switch logic from Factory.Create
3. Deprecate Factory type
4. Update main.go callers

**Verify**: `make test && make lint`

### Phase 4: ChainContext Resolver

1. Add `Resolve(key string)` method to ChainContext
2. Refactor `ExpandVariables` to use vars.ExpandString
3. Ensure resolution priority maintained

**Verify**: `make test && make lint`

### Phase 5: Main.go Extraction

1. Create `internal/output/result.go` with PrintJSON
2. Create `internal/runner/stress.go` with RunStress
3. Create `internal/importer/cli.go` with RunImport
4. Update main.go to call new packages

**Verify**: `make test && make lint`

## Verification Checklist

- [ ] All tests pass: `make test`
- [ ] Linter passes: `make lint`
- [ ] main.go is smaller (check with `wc -l cmd/yapi/main.go`)
- [ ] LSP still works: test in VS Code
- [ ] CLI still works: `yapi run examples/http.yapi.yml`

## Rollback

If issues arise, each phase is independent. Revert phase-specific commits.
