# Feature Specification: API Consolidation Refactor

**Branch**: `002-api-consolidation-refactor` | **Created**: 2026-01-03 | **Status**: Draft

## Overview

Refactor yapi's internal codebase to remove dead code, consolidate the analyzer API into a single entry point, extract shared AST logic, deduplicate variable expansion logic, and reduce main.go complexity by extracting output formatting, stress testing, and import handling into dedicated packages.

## User Stories

### US1 - Maintainer Simplifies Validation API (P1)

A yapi maintainer wants to analyze config files using a single, consistent API instead of multiple combinatorial functions with different parameter signatures.

**CLI Usage**:
```bash
yapi validate file.yapi.yml
yapi run file.yapi.yml
```

**Acceptance**:
- Given a yapi config file, when the validation system analyzes it, then only one entry point function is called regardless of project context or environment settings
- Given any combination of project config, environment, and strictness settings, when calling the analyzer, then a single `Analyze()` function handles all scenarios via an options struct

---

### US2 - Maintainer Uses Accurate Diagnostic Line Numbers (P2)

A yapi maintainer or IDE integration wants validation diagnostics to report accurate line numbers instead of defaulting to line 0 for env file and project-level errors.

**CLI Usage**:
```bash
yapi validate file.yapi.yml
```

**Acceptance**:
- Given a config file with a missing env file reference, when validation runs, then the diagnostic reports the actual line number where env_files is declared
- Given a project config validation error, when the LSP reports it, then the line number points to the exact location in yapi.config.yml

---

### US3 - Maintainer Extends Transport Support (P2)

A yapi maintainer wants to add or modify transport types without navigating a Factory struct pattern.

**CLI Usage**:
```bash
yapi run http-request.yapi.yml
yapi run grpc-request.yapi.yml
```

**Acceptance**:
- Given a request with any transport type, when the executor runs, then a simple `GetTransport()` function returns the appropriate transport function
- Given a new transport type to add, when a developer implements it, then they only need to add a case to a single switch statement

## Requirements

### Functional

- **FR-001**: `yapi` MUST provide a single `validation.Analyze(text string, opts AnalyzeOptions)` function as the entry point for all config analysis
- **FR-002**: ~~DROPPED~~ (Research found these structs are actively used)
- **FR-003**: `yapi` MUST replace the `Factory` struct in executor with a standalone `GetTransport(transport string, client HTTPClient)` function
- **FR-004**: `yapi` MUST extract YAML AST position-finding logic into `internal/validation/ast.go` with exported `FindVarPositionInYAML()` function
- **FR-005**: `yapi` MUST report accurate line numbers for env file validation errors instead of line 0
- **FR-006**: `yapi` MUST use the compiler package for request validation instead of duplicating validation logic
- **FR-007**: `yapi` MUST have `ChainContext` implement `vars.Resolver` interface and delegate variable expansion to the vars package
- **FR-008**: `yapi` MUST extract JSON output formatting to `internal/output/result.go`
- **FR-009**: `yapi` MUST extract stress test logic to `internal/runner/stress.go`
- **FR-010**: `yapi` MUST extract import CLI handler to `internal/importer/cli.go`

### Code Organization

| Package                            | Changes                                    |
|------------------------------------|--------------------------------------------|
| `internal/config/loader.go`        | No changes (structs are used)              |
| `internal/validation/analyzer.go`  | Consolidate to single `Analyze()` function |
| `internal/validation/ast.go`       | New file with shared AST helpers           |
| `internal/executor/executor.go`    | Replace Factory with `GetTransport()`      |
| `internal/langserver/langserver.go`| Use shared AST helpers                     |
| `internal/runner/context.go`       | Implement `vars.Resolver`                  |
| `internal/runner/stress.go`        | New file for stress test logic             |
| `internal/output/result.go`        | New file for JSON output                   |
| `internal/importer/cli.go`         | New file for import handler                |
| `cmd/yapi/main.go`                 | Reduced size, delegates to new packages    |

### Protocol Support

| Protocol | Supported | Notes                |
|----------|-----------|----------------------|
| HTTP     | [x]       | Via `GetTransport()` |
| gRPC     | [x]       | Via `GetTransport()` |
| GraphQL  | [x]       | Via `GetTransport()` |
| TCP      | [x]       | Via `GetTransport()` |

## Edge Cases

- What happens when `FindVarPositionInYAML` cannot locate a key path? Returns nil Location, caller falls back to line 0
- What happens when project config has no environments defined? `Analyze()` proceeds without environment defaults
- How does `ChainContext.Resolve()` handle non-existent variables? Returns empty string without error (existing behavior preserved)
- What happens when a transport type is unrecognized? `GetTransport()` returns an error with clear message

## Assumptions

- ~~The dead code structs were assumed unused~~ **Research found these are actively used - FR-002 dropped**
- The existing `AnalyzeConfigString*` function variants all share the same core logic that can be unified
- The LSP and validation packages can share AST helper code without circular dependencies
- Extracting code to new packages maintains the same public API behavior for CLI commands
- Tests exist and will verify that refactoring does not break existing functionality

## Success Criteria

- [ ] Single entry point `Analyze()` function handles all validation scenarios
- [x] ~~Dead code removed~~ FR-002 dropped (structs are used)
- [ ] Executor Factory replaced with standalone function
- [ ] AST helpers consolidated in validation/ast.go
- [ ] Diagnostic line numbers accurate for env file errors (not line 0)
- [ ] ChainContext uses vars package for variable expansion
- [ ] main.go reduced in size by extracting output, stress, and import logic
- [ ] Tests pass: `make test`
- [ ] Lint passes: `make lint`
- [ ] No changes to CLI command signatures or user-facing behavior
- [ ] There is more code deleted from the codebase than added
