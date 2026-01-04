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

## State Transitions

### Env File Loading State Machine

```
┌─────────────┐
│   START     │
└──────┬──────┘
       │
       ▼
┌─────────────┐     file missing     ┌─────────────┐
│ Check File  │─────────────────────►│  WARNING    │
│   Exists    │                      │ (continue)  │
└──────┬──────┘                      └─────────────┘
       │ exists
       ▼
┌─────────────┐   permission denied  ┌─────────────┐
│ Check File  │─────────────────────►│   ERROR     │
│  Readable   │                      │   (halt)    │
└──────┬──────┘                      └─────────────┘
       │ readable
       ▼
┌─────────────┐
│  Parse Env  │
│    File     │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   LOADED    │
└─────────────┘
```

### Strict Mode Behavior

| Condition          | Default Mode        | Strict Mode         |
|--------------------|---------------------|---------------------|
| File missing       | Warning, continue   | Error, halt         |
| Permission denied  | Error, halt         | Error, halt         |
| Parse error        | Error, halt         | Error, halt         |
| File valid         | Load variables      | Load variables      |

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
