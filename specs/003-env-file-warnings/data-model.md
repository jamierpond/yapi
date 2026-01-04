# Data Model: Env File Warnings

**Feature**: 001-env-file-warnings
**Date**: 2026-01-03

## Overview

This feature does not introduce new persistent entities. It extends existing in-memory structures to track env file validation state.

## Extended Structures

### EnvFileStatus

New structure to track env file validation state during config loading.

```go
// EnvFileStatus represents the validation state of an env file reference
type EnvFileStatus struct {
    Path       string    // Original path from config
    Resolved   string    // Absolute path after resolution
    Exists     bool      // Whether file exists
    Readable   bool      // Whether file is readable (if exists)
    Error      error     // Error if not readable (permission denied, etc.)
    Line       int       // Line number in source YAML
    Col        int       // Column number in source YAML
}
```

### EnvFileLoadResult

Extended return type for env file loading.

```go
// EnvFileLoadResult contains the result of loading env files
type EnvFileLoadResult struct {
    Variables   map[string]string   // Merged variables from all valid files
    Warnings    []string            // Warnings for missing files
    FileStatus  []EnvFileStatus     // Status of each env file
}
```

## Existing Structures (Unchanged)

### ConfigV1.EnvFiles

```yaml
env_files:
  - .env.local
  - .env.secrets
```

```go
type ConfigV1 struct {
    // ...existing fields...
    EnvFiles []string `yaml:"env_files,omitempty"`
}
```

No schema changes required.

### validation.Diagnostic

Existing diagnostic structure, used for LSP diagnostics.

```go
type Diagnostic struct {
    Severity Severity
    Field    string
    Message  string
    Line     int
    Col      int
}
```

## Variable Resolution

### Resolution Priority

```
Default mode:       env_files > OS env (warning on fallback, info shows source)
--strict-env mode:  env_files only (no OS fallback)
```

### Resolution Flow

```
┌─────────────────┐
│  Resolve ${VAR} │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     found      ┌─────────────────┐
│ Check env_files │───────────────►│ Use env_files   │
└────────┬────────┘                │ value           │
         │ not found               └─────────────────┘
         ▼
┌─────────────────┐
│ --strict-env?   │
└────────┬────────┘
    yes  │    no
    ▼    │    ▼
┌────────┴────┐  ┌─────────────────┐     found     ┌─────────────────┐
│   ERROR:    │  │ Check OS env    │──────────────►│ Use OS value    │
│  undefined  │  └────────┬────────┘               │ + emit WARNING  │
└─────────────┘           │ not found              └─────────────────┘
                          ▼
                 ┌─────────────────┐
                 │ ERROR: undefined│
                 └─────────────────┘
```

### Env File Loading

| Condition          | Default Mode          | --strict-env Mode   |
|--------------------|-----------------------|---------------------|
| File missing       | Warning, skip file    | Error, stop         |
| Permission denied  | Error, stop           | Error, stop         |
| Parse error        | Error, stop           | Error, stop         |
| File valid         | Load variables        | Load variables      |
| Var from OS env    | Warning, use value    | Error (no fallback) |

## Relationships

```
ConfigV1
    └── env_files: []string
            │
            ├── resolves to ──► EnvFileStatus (per file)
            │                        │
            │                        ├── Exists = true ──► Parse & Load
            │                        │
            │                        └── Exists = false ──► Warning/Error
            │
            └── produces ──► EnvFileLoadResult
                                │
                                ├── Variables (merged)
                                ├── Warnings (missing files)
                                └── FileStatus (for diagnostics)
```

## LSP Document Context

Existing `document` struct gains no new fields. Env file status is computed on-demand during validation.

```go
type document struct {
    URI         protocol.DocumentUri
    Text        string
    ProjectRoot string
    Project     *config.ProjectConfigV1
    // Env file status computed during validateAndNotify()
}
```
