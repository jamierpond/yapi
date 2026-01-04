# Quickstart: Env File Warnings

**Feature**: 001-env-file-warnings
**Date**: 2026-01-03

## Quick Reference

### CLI Usage

```bash
# Default: env_files take priority, fall back to OS env with warning
yapi run my-request.yapi.yml

# Strict mode: env_files only, no OS env fallback, error on missing files
yapi run my-request.yapi.yml --strict-env
```

### Example YAPI Config

```yaml
yapi: v1
url: https://api.github.com/repos/user/repo
method: GET
env_files:
  - .env.local    # Warning if missing
  - .env.secrets  # Warning if missing
headers:
  Authorization: Bearer ${GITHUB_PAT}  # Diagnostic if undefined
```

### Expected Output

**Missing file + OS env fallback (default mode):**
```
Warning: env file '.env.local' not found
Warning: variable 'GITHUB_PAT' resolved from OS environment, not configuration
{"stars": 42, "forks": 10, "name": "user/repo"}
```

**Missing file (strict mode):**
```
Error: env file '.env.local' not found
exit status 1
```

**Undefined variable (strict mode):**
```
Error: variable 'GITHUB_PAT' not defined in env_files
exit status 1
```

### LSP Features

1. **Go to Definition on env file path**: Click on `.env.local` → opens file at line 1
2. **Go to Definition on variable**: Click on `${GITHUB_PAT}` → opens .env file at variable definition
3. **Diagnostics**: Squiggly underlines on missing files and undefined variables

## Implementation Entry Points

| Feature                     | File                              | Function/Method                |
|-----------------------------|-----------------------------------|--------------------------------|
| Warning on missing file     | `internal/config/loader.go`       | `loadEnvFiles()`               |
| `--strict-env` flag         | `cmd/yapi/main.go`                | Run command flags              |
| LSP go-to-definition        | `internal/langserver/langserver.go` | `textDocumentDefinition()`    |
| LSP diagnostics             | `internal/validation/analyzer.go` | `AnalyzeConfigString()`        |

## Test Commands

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/config/... -v
go test ./internal/langserver/... -v
go test ./internal/validation/... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Lint
make lint
```

## Manual Testing

1. **CLI Warning Test**:
   ```bash
   # Create a YAPI file referencing non-existent .env
   echo 'yapi: v1
   url: https://httpbin.org/get
   method: GET
   env_files:
     - .env.nonexistent' > test.yapi.yml

   # Should show warning but complete
   yapi run test.yapi.yml

   # Should fail with error
   yapi run test.yapi.yml --strict-env
   ```

2. **LSP Test in VS Code**:
   - Open a .yapi.yml file with env_files
   - Verify squiggly underline on missing files
   - Ctrl+click on env file path → should navigate
   - Ctrl+click on `${VAR}` → should navigate to definition
