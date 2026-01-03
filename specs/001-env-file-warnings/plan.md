# Implementation Plan: Env File Warnings

**Branch**: `001-env-file-warnings` | **Date**: 2026-01-03 | **Spec**: [spec.md](./spec.md)

## Summary

Add warnings for missing env files in CLI execution (with optional strict mode to treat as errors), plus comprehensive LSP support for diagnostics and go-to-definition on env file paths and `${VAR}` references.

## Technical Context

**Language**: Go 1.21+
**Key Packages**:
- `internal/config/loader.go` - env file loading (loadEnvFiles)
- `internal/langserver/langserver.go` - LSP server, go-to-definition
- `internal/validation/analyzer.go` - FindEnvVarRefs, diagnostics
- `cmd/yapi/main.go` - CLI entry point, flag handling

**Testing**: `go test`, table-driven tests
**Build**: `make build && make test && make lint`

## Constitution Check

*Must pass before implementation.*

| Principle        | Status | Notes                                                    |
|------------------|--------|----------------------------------------------------------|
| CLI-First        | [x]    | `--strict-env` flag, warnings to stderr                  |
| Git-Friendly     | [x]    | No changes to YAML schema, existing env_files field      |
| Protocol Agnostic| [x]    | Env files work across all protocols                      |
| Simplicity       | [x]    | Minimal additions, leverages existing infrastructure     |
| Dogfooding       | [x]    | Webapp can use yapi with env_files                       |

## Affected Areas

```text
cmd/yapi/
├── main.go              # Add --strict-env flag to run command

internal/
├── config/
│   └── loader.go        # Modify loadEnvFiles to warn instead of error
├── langserver/
│   └── langserver.go    # Add go-to-definition for env_files entries
└── validation/
    └── analyzer.go      # Add env file existence diagnostics
```

## Implementation Approach

1. **CLI Warning Behavior**: Modify `loadEnvFiles()` to return warnings for missing files instead of errors, continue loading remaining files. Add `--strict-env` flag to `run` command that converts warnings to errors.

2. **LSP Go-to-Definition for env_files**: Extend `textDocumentDefinition()` to detect cursor on env file paths in `env_files` array. Return location pointing to line 1 of the file.

3. **LSP Diagnostics for Missing Files**: Add validation in `AnalyzeConfigString()` to check if each `env_files` entry exists. Generate warning diagnostics with file path and line/column position.

4. **LSP Diagnostics for Undefined Variables**: Already partially implemented via `FindEnvVarRefs()` and `ValidateProjectVars()`. Ensure diagnostics appear for variables not defined in any env file.

## Phase Implementation

### Phase 1: CLI Warnings (FR-001 through FR-006)

**Files to modify:**
- `internal/config/loader.go`: Change `loadEnvFiles()` to:
  - Return `(map[string]string, []string, error)` - vars, warnings, error
  - On missing file: add warning, continue to next file
  - On permission error: return error (halt)
- `cmd/yapi/main.go`: Add `--strict-env` flag to `run` command
- Update callers of `loadEnvFiles()` to handle warnings

### Phase 2: LSP Go-to-Definition (FR-007 through FR-010)

**Files to modify:**
- `internal/langserver/langserver.go`:
  - Add `findEnvFilePathAtPosition()` helper to detect cursor on env_files entries
  - Extend `textDocumentDefinition()` to handle env file paths
  - Return `protocol.Location` with line 1 of target file
  - For `${VAR}` references: existing code already works, verify edge cases

### Phase 3: LSP Diagnostics (FR-011, FR-012)

**Files to modify:**
- `internal/validation/analyzer.go`:
  - Add `validateEnvFilesExist()` function
  - Called during `AnalyzeConfigString()` / `AnalyzeConfigStringWithProject()`
  - Returns diagnostics for missing files with line/column from YAML position
- Undefined variable diagnostics: verify existing `ValidateProjectVars()` coverage

## Complexity Justification

No complexity concerns. Implementation leverages existing patterns:
- Warning collection pattern already used in loader.go
- Go-to-definition pattern exists for `${VAR}` references
- Diagnostic publishing pattern well-established in langserver.go

## Test Strategy

| Test Case                              | Type        | Location                           |
|----------------------------------------|-------------|------------------------------------|
| Missing env file shows warning         | Unit        | `internal/config/loader_test.go`   |
| --strict-env exits on missing file     | Integration | `cmd/yapi/main_test.go`            |
| Permission error always halts          | Unit        | `internal/config/loader_test.go`   |
| LSP go-to-definition on env file path  | Unit        | `internal/langserver/langserver_test.go` |
| LSP diagnostic for missing env file    | Unit        | `internal/validation/analyzer_test.go` |
| LSP diagnostic for undefined variable  | Unit        | `internal/validation/analyzer_test.go` |

## Success Metrics

- `make test` passes with new test coverage
- `make lint` passes
- Manual verification: open YAPI file with missing .env, see warning underline
- Manual verification: Ctrl+click on `.env.local` navigates to file
- Manual verification: Ctrl+click on `${VAR}` navigates to definition line
