# Data Model: API Consolidation Refactor

**Date**: 2026-01-03
**Feature**: 002-api-consolidation-refactor

## Overview

This refactoring feature introduces minimal new types. The primary changes are API consolidation and code organization.

## New Types

### 1. AnalyzeOptions (validation package)

Consolidates parameters for the analyzer API.

```go
type AnalyzeOptions struct {
    FilePath    string                  // Path to config file for relative resolution
    Project     *config.ProjectConfigV1 // Project context (optional)
    ProjectRoot string                  // Root directory of the project (optional)
    StrictEnv   bool                    // Strict environment mode (optional)
}
```

**Validation Rules**:
- All fields are optional
- If `Project` is set, `ProjectRoot` should also be set
- `FilePath` used for resolving relative paths in includes

### 2. Location (validation package)

Protocol-agnostic position type for AST helpers.

```go
type Location struct {
    File string // Absolute file path
    Line int    // 0-indexed line number
    Col  int    // 0-indexed column number
}
```

**Validation Rules**:
- `File` must be an absolute path
- `Line` and `Col` are 0-indexed for internal use
- Callers (LSP) convert to 1-indexed as needed

## Modified Types

### 1. ChainContext (runner package)

Existing type gains new interface implementation.

```go
// ChainContext now implements vars.Resolver
type ChainContext struct {
    Results      map[string]StepResult
    EnvOverrides map[string]string
}

// New method implementing vars.Resolver
func (c *ChainContext) Resolve(key string) (string, error)
```

**Resolution Priority** (unchanged):
1. OS Environment variables
2. EnvOverrides from project config
3. Chain results (for keys containing dots)

## Deprecated Types

### 1. Factory (executor package)

```go
// Deprecated: Use GetTransport() instead
type Factory struct {
    Client HTTPClient
}
```

**Migration Path**: Replace `NewFactory(client).Create(transport)` with `GetTransport(transport, client)`

## Type Relationships

```
┌─────────────────────────────────────────────────────────┐
│                    validation package                    │
├─────────────────────────────────────────────────────────┤
│  AnalyzeOptions ──uses──> config.ProjectConfigV1        │
│  Analyze() ──returns──> *Analysis                       │
│  FindVarPositionInYAML() ──returns──> *Location         │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                   langserver package                     │
├─────────────────────────────────────────────────────────┤
│  ──calls──> validation.FindVarPositionInYAML()          │
│  ──converts──> Location → protocol.Location             │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                     runner package                       │
├─────────────────────────────────────────────────────────┤
│  ChainContext ──implements──> vars.Resolver             │
│  ExpandVariables() ──delegates──> vars.ExpandString()   │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                    executor package                      │
├─────────────────────────────────────────────────────────┤
│  GetTransport() ──returns──> TransportFunc              │
│  Factory (deprecated) ──wraps──> GetTransport()         │
└─────────────────────────────────────────────────────────┘
```

## State Transitions

N/A - This refactoring does not introduce stateful types.
