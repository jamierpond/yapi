# Implementation Plan: API Consolidation Refactor

**Branch**: `002-api-consolidation-refactor` | **Date**: 2026-01-03 | **Spec**: [spec.md](./spec.md)

## Summary

Consolidate yapi's internal APIs by: (1) unifying the 4-function analyzer chain into single `Analyze()` with options struct, (2) replacing executor Factory with standalone `GetTransport()`, (3) extracting shared AST helpers from langserver to validation package, (4) having ChainContext implement `vars.Resolver`, and (5) extracting JSON output, stress testing, and import logic from main.go.

## Technical Context

**Language**: Go 1.21+
**Key Packages**: `internal/validation`, `internal/executor`, `internal/langserver`, `internal/runner`, `internal/vars`, `internal/compiler`
**Testing**: `go test`, table-driven tests
**Build**: `make build && make test && make lint`

## Constitution Check

*Must pass before implementation.*

| Principle | Status | Notes |
|-----------|--------|-------|
| CLI-First | [x] | No CLI changes - internal refactoring only |
| Git-Friendly | [x] | No config format changes |
| Protocol Agnostic | [x] | GetTransport() maintains protocol parity |
| Simplicity | [x] | Reduces function count, consolidates patterns |
| Dogfooding | [x] | LSP uses same code paths as CLI |
| Minimal Code | [x] | Target: net reduction in LOC |
| Single Code Path | [x] | Core motivation - LSP/CLI share validation |

## Affected Areas

```text
cmd/yapi/main.go           # Extract JSON output, stress, import (~400 lines)
internal/
├── validation/
│   ├── analyzer.go        # Consolidate 4 functions → 1
│   └── ast.go             # NEW: shared AST position helpers
├── executor/executor.go   # Factory → GetTransport()
├── langserver/langserver.go # Use validation.ast helpers
├── runner/
│   ├── context.go         # Implement vars.Resolver
│   └── stress.go          # NEW: extracted from main.go
├── output/result.go       # NEW: JSON output formatting
└── importer/cli.go        # NEW: import command handler
```

## Implementation Approach

### Phase 1: AST Extraction & Line Number Fix (FR-004, FR-005)

1. Create `internal/validation/ast.go` with `Location` type and exported functions:
   - `FindVarPositionInYAML(rootPath, fileRelPath string, section []string) (*Location, error)`
   - `FindNodeInMapping(node *yaml.Node, key string) *yaml.Node`
   - `FindKeyNodeInMapping(node *yaml.Node, key string) *yaml.Node`

2. Move implementations from `langserver.go` (lines 787, 906, 922) to `ast.go`

3. Update `langserver.go` to import and call `validation.FindVarPositionInYAML()`

4. Update `ValidateEnvFilesExistFromProject` in `analyzer.go` to use AST helpers for accurate line numbers

### Phase 2: Analyzer Consolidation (FR-001)

1. Define `AnalyzeOptions` struct in `analyzer.go`:
   ```go
   type AnalyzeOptions struct {
       FilePath    string
       Project     *config.ProjectConfigV1
       ProjectRoot string
       StrictEnv   bool
   }
   ```

2. Create new `Analyze(text string, opts AnalyzeOptions) (*Analysis, error)` function

3. Refactor internal `analyzeParsed` to accept `AnalyzeOptions`

4. Deprecate existing functions (keep as wrappers for backward compatibility):
   - `AnalyzeConfigString` → calls `Analyze(text, AnalyzeOptions{})`
   - `AnalyzeConfigStringWithProject` → calls `Analyze(text, AnalyzeOptions{Project: p, ProjectRoot: r})`
   - etc.

### Phase 3: Executor Simplification (FR-003)

1. Add standalone function in `executor.go`:
   ```go
   func GetTransport(transport string, client HTTPClient) (TransportFunc, error)
   ```

2. Move switch logic from `Factory.Create()` to `GetTransport()`

3. Mark `Factory` and `NewFactory` as deprecated

4. Update all callers in `main.go` to use `GetTransport()` directly

### Phase 4: ChainContext as Resolver (FR-006, FR-007)

1. Add `Resolve(key string) (string, error)` method to `ChainContext` implementing `vars.Resolver`

2. Refactor `ExpandVariables` to delegate to `vars.ExpandString(input, c.Resolve)`

3. Ensure resolution priority maintained: OS env > EnvOverrides > Chain results

### Phase 5: Main.go Extraction (FR-008, FR-009, FR-010)

1. **Extract JSON output** to `internal/output/result.go`:
   - Move `printResultAsJSON` function and related structs
   - Export as `output.PrintJSON(result, chain, err)`

2. **Extract stress testing** to `internal/runner/stress.go`:
   - Move worker pool logic from main.go
   - Export as `runner.RunStress(path, concurrency, requests)`

3. **Extract import handling** to `internal/importer/cli.go`:
   - Move Postman import CLI handler
   - Export as `importer.RunImport(cmd, args)`

4. Update main.go to call new package functions

## Complexity Justification

| Concern | Why Needed | Simpler Alternative Rejected |
|---------|------------|------------------------------|
| AST extraction creates new file | Enables shared position-finding for LSP and validation | Duplicate code in both packages (violates VII) |
| Deprecation wrappers | Maintains backward compatibility | Breaking change to public API |

## Testing Strategy

1. **Regression tests**: Ensure existing tests pass after each phase
2. **New tests for AST helpers**: Position-finding accuracy tests
3. **Integration tests**: Verify LSP and CLI produce identical diagnostics
4. **Line counting**: Verify net LOC reduction after all phases

## Success Metrics

- [ ] `make test` passes after each phase
- [ ] `make lint` passes after each phase
- [ ] main.go reduced by ~400 lines
- [ ] No duplicate validation/AST logic between langserver and validation
- [ ] Net LOC reduction (more deleted than added)
