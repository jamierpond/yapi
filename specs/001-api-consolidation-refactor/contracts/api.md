# API Contracts: API Consolidation Refactor

**Date**: 2026-01-03
**Feature**: 002-api-consolidation-refactor

## Overview

This is an internal refactoring feature. No external API contracts (REST, GraphQL, etc.) are affected. This document describes the internal Go API contracts.

## Internal API Contracts

### 1. validation.Analyze

**New Entry Point** (replaces 4 existing functions)

```go
// Analyze is the single entry point for config analysis.
// All optional parameters are provided via AnalyzeOptions.
func Analyze(text string, opts AnalyzeOptions) (*Analysis, error)

type AnalyzeOptions struct {
    FilePath    string                  // optional
    Project     *config.ProjectConfigV1 // optional
    ProjectRoot string                  // optional
    StrictEnv   bool                    // optional, default false
}
```

**Behavior Contract**:
- If `opts` is zero-valued, behaves like current `AnalyzeConfigString(text)`
- If `Project` is set, applies environment defaults from project config
- If `StrictEnv` is true, treats unresolved variables as errors
- Returns `*Analysis` with diagnostics, never returns error for parse failures (wraps as diagnostic)

### 2. validation.FindVarPositionInYAML

**New AST Helper**

```go
// FindVarPositionInYAML finds the position of a key path in a YAML file.
// Returns nil if path not found.
func FindVarPositionInYAML(rootPath, fileRelPath string, section []string) (*Location, error)

type Location struct {
    File string // absolute path
    Line int    // 0-indexed
    Col  int    // 0-indexed
}
```

**Behavior Contract**:
- Reads file at `filepath.Join(rootPath, fileRelPath)`
- Navigates YAML tree following `section` path (e.g., `["environments", "dev", "vars"]`)
- Returns nil Location if path not found (not an error)
- Returns error only for file read or parse failures

### 3. executor.GetTransport

**New Entry Point** (replaces Factory)

```go
// GetTransport creates the appropriate transport function.
// Wraps transport with timing middleware.
func GetTransport(transport string, client HTTPClient) (TransportFunc, error)
```

**Behavior Contract**:
- Accepts transport types: "http", "graphql", "grpc", "tcp"
- Returns error for unsupported transport type
- Always wraps result with `WithTiming` middleware
- `client` is used for HTTP and GraphQL transports only

### 4. ChainContext.Resolve

**New Interface Implementation**

```go
// Resolve implements vars.Resolver.
// Resolution priority: OS env > EnvOverrides > Chain results
func (c *ChainContext) Resolve(key string) (string, error)
```

**Behavior Contract**:
- Returns OS environment variable if exists
- Falls back to EnvOverrides if key exists
- Falls back to chain results for keys containing "." (e.g., "step1.body")
- Returns empty string (not error) for unresolved keys

### 5. output.PrintJSON

**New Export**

```go
// PrintJSON formats and prints execution result as JSON.
func PrintJSON(result *runner.Result, chain *runner.ChainResult, err error) error
```

**Behavior Contract**:
- Marshals result to JSON and prints to stdout
- Includes error information if `err` is non-nil
- Returns error only if JSON marshaling fails

### 6. runner.RunStress

**New Export**

```go
// RunStress executes stress testing with worker pool.
func RunStress(path string, concurrency int, requests int) error
```

**Behavior Contract**:
- Creates `concurrency` workers
- Each worker executes up to `requests` total
- Returns first fatal error encountered
- Prints statistics on completion

### 7. importer.RunImport

**New Export**

```go
// RunImport handles the import CLI command.
func RunImport(cmd *cobra.Command, args []string) error
```

**Behavior Contract**:
- Detects import format (Postman, cURL, etc.) from input
- Writes converted .yapi.yml file(s)
- Returns error with actionable message on failure

## Deprecated APIs

### Deprecated Functions

```go
// Deprecated: Use Analyze() instead
func AnalyzeConfigString(text string) (*Analysis, error)
func AnalyzeConfigStringWithProject(text string, project *config.ProjectConfigV1, projectRoot string) (*Analysis, error)
func AnalyzeConfigStringWithProjectAndPath(text string, configPath string, project *config.ProjectConfigV1, projectRoot string) (*Analysis, error)
func AnalyzeConfigStringWithProjectAndPathAndOptions(text string, configPath string, project *config.ProjectConfigV1, projectRoot string, opts AnalyzeOptions) (*Analysis, error)

// Deprecated: Use GetTransport() instead
type Factory struct
func NewFactory(client HTTPClient) *Factory
func (f *Factory) Create(transport string) (TransportFunc, error)
```

**Migration Timeline**: Deprecated in this release, removal in next major version.
