# Research: Env File Warnings

**Feature**: 001-env-file-warnings
**Date**: 2026-01-03

## Technical Decisions

### 1. Variable Resolution Priority

**Decision**: Config takes priority over OS env. Order: `env.[name].vars` > `env.[name].env_files` > `defaults.vars` > `defaults.env_files` > per-request `env_files` > OS env (fallback with info diagnostic).

**Rationale**:
- Config should be explicit and reproducible
- OS env as fallback catches "works on my machine" issues via info diagnostic
- `--strict-env` disables OS fallback entirely for CI reproducibility

**Alternatives Considered**:
- OS env takes priority: Current behavior, but hides config issues
- No OS fallback ever: Too strict for local development

### 2. Warning vs Error Behavior for Missing Env Files

**Decision**: Default to warning (non-blocking), with `--strict-env` flag for error behavior.

**Rationale**:
- Matches common tool behavior (dotenv, docker-compose warn on missing files)
- Allows development workflows where not all env files exist locally
- Strict mode provides CI/production safety net

**Alternatives Considered**:
- Always error: Too disruptive for local development
- Config-level flag in YAML: Over-engineering for simple use case

### 3. Permission Error Handling

**Decision**: Permission errors are always errors (not warnings), regardless of strict mode.

**Rationale**:
- File exists but unreadable indicates system misconfiguration
- Different from "file doesn't exist" case
- User needs immediate feedback to fix permissions

**Alternatives Considered**:
- Treat as warning: Could hide real problems silently

### 4. LSP Go-to-Definition Target Position

**Decision**: Navigate to line 1, column 0 of the target env file.

**Rationale**:
- Env files have no specific "declaration" line for the file itself
- Line 1 is the natural starting point
- Consistent with how IDEs handle file navigation

**Alternatives Considered**:
- Navigate to first non-comment line: Adds complexity, minimal benefit

### 5. Multiple Env File Definitions

**Decision**: When a variable is defined in multiple env files, go-to-definition navigates to the first definition per `env_files` array order.

**Rationale**:
- Matches runtime precedence (later files override earlier)
- Predictable behavior for users
- Already implemented in existing go-to-definition code

**Alternatives Considered**:
- Show picker with all definitions: Adds complexity, LSP support varies by editor

### 6. Env File Path Detection in LSP

**Decision**: Use YAML node positions to detect cursor on env_files array entries.

**Rationale**:
- Leverages existing YAML parsing infrastructure
- Provides accurate line/column positions for diagnostics
- Handles quoted and unquoted strings correctly

**Alternatives Considered**:
- Regex-based detection: Fragile, doesn't handle edge cases

## Existing Code Patterns

### Warning Collection Pattern

From `internal/config/loader.go`:
```go
type ParseResult struct {
    Request  *domain.Request
    Warnings []string  // Already exists, used for "missing yapi: v1"
    // ...
}
```

This pattern can be extended for env file warnings.

### Go-to-Definition Pattern

From `internal/langserver/langserver.go`:
```go
func textDocumentDefinition(ctx *glsp.Context, params *protocol.DefinitionParams) (any, error) {
    // Find reference at cursor position
    // Locate definition
    // Return protocol.Location
}
```

Extend this to also check for env file path positions.

### Diagnostic Publishing Pattern

From `internal/langserver/langserver.go`:
```go
ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
    URI:         uri,
    Diagnostics: diagnostics,
})
```

Use same pattern for env file existence diagnostics.

## Dependencies

- `github.com/joho/godotenv`: Already used for .env file parsing
- `gopkg.in/yaml.v3`: Already used for YAML parsing with node positions
- `github.com/tliron/glsp`: Already used for LSP protocol

No new dependencies required.

## Performance Considerations

### LSP File Existence Checks

**Risk**: Checking file existence on every keystroke could be slow.

**Mitigation**:
- File existence checks are already performed during document analysis
- Results are cached in document context
- Only check when document is opened/changed, not on every operation

### Env File Parsing

**Risk**: Parsing multiple env files on every validation.

**Mitigation**:
- `ProjectConfigV1.envCache` already caches resolved env files
- Cache invalidation happens on document change
- Env files are typically small (< 1KB)

### 7. Diagnostic Severity for Variable Resolution

**Decision**:
- **Warning** when variable falls back to OS env (potential reproducibility issue)
- **Info** to show where each variable was resolved from (source transparency)

**Rationale**:
- OS fallback is a warning because it's a potential "works on my machine" issue
- Info diagnostic provides transparency about resolution source for all variables
- Users can see at a glance where their variables come from

**Alternatives Considered**:
- Info for OS fallback: Doesn't emphasize the reproducibility risk enough
- No source info: Users can't easily understand variable resolution

## Edge Cases Covered

| Edge Case | Handling |
|-----------|----------|
| Empty env_files array | No warnings, proceed normally |
| Absolute vs relative paths | Both supported, relative resolved from YAML file |
| Duplicate entries | Warn/check once per unique path |
| Empty env file | Valid, no warning |
| Windows paths | OS-appropriate path handling via filepath package |
| Env file with syntax errors | godotenv.Read returns error, show as diagnostic |
| Multi-environment project | Each env can have own env_files, defaults apply to all |
| Variable in OS but not config | Info diagnostic, use OS value (unless --strict-env) |
