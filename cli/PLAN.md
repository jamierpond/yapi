# Plan: Integrated Server Startup for yapi test

## Problem Statement

Currently, to run yapi tests against a local server, users must:
1. Manually start their dev server in one terminal
2. Wait for it to be healthy
3. Run `yapi test` in another terminal
4. Remember to stop the server

The GitHub Action solves this for CI, but local development requires this manual workflow.

## Solution

Add a `test` block to `yapi.config.yml` that automatically handles server lifecycle:

```yaml
yapi: v1

test:
  start: "npm run dev"
  wait_on:
    - "http://localhost:3000/healthz"
    - "grpc://localhost:50051"
  timeout: 60s
  parallel: 8

environments:
  local:
    url: http://localhost:3000
```

## Command Behavior

```bash
# Normal usage - starts server, waits, runs tests, kills server
yapi test ./tests

# Skip server (already running)
yapi test ./tests --no-start

# Override from CLI
yapi test ./tests --start "npm run dev:debug" --wait-on "http://localhost:4000/health"

# Verbose - see server output
yapi test ./tests --verbose
```

## Health Check Protocol Support

| Protocol | URL Format | Behavior |
|----------|------------|----------|
| HTTP/HTTPS | `http://localhost:3000/healthz` | Poll until 2xx response |
| gRPC | `grpc://localhost:50051` | Use `grpc.health.v1.Health/Check` |
| TCP | `tcp://localhost:5432` | Poll until connection succeeds |

## Config Schema

```yaml
test:
  # Server startup (optional - if not present, tests run without starting anything)
  start: "npm run dev"

  # Health check URL(s) - supports single string or array
  wait_on: "http://localhost:3000/healthz"
  # OR
  wait_on:
    - "http://localhost:3000/healthz"
    - "tcp://localhost:5432"

  # How long to wait for health checks (default: 60s)
  timeout: 60s

  # Test execution defaults (optional - can still be overridden via CLI flags)
  parallel: 4
  directory: "./tests"
  verbose: false
  all: false
```

## Implementation Plan

### 1. Config Schema Changes

**File:** `cli/internal/config/project.go`

```go
type TestConfig struct {
    Start     string        `yaml:"start,omitempty"`
    WaitOn    StringOrArray `yaml:"wait_on,omitempty"`
    Timeout   string        `yaml:"timeout,omitempty"`
    Parallel  int           `yaml:"parallel,omitempty"`
    Directory string        `yaml:"directory,omitempty"`
    Verbose   bool          `yaml:"verbose,omitempty"`
    All       bool          `yaml:"all,omitempty"`
}

type ProjectConfigV1 struct {
    // ... existing fields ...
    Test *TestConfig `yaml:"test,omitempty"`
}
```

### 2. Health Check Implementation

**New file:** `cli/internal/healthcheck/healthcheck.go`

```go
// WaitForHealth polls all URLs until they're healthy or timeout
func WaitForHealth(ctx context.Context, urls []string, timeout time.Duration) error {
    // Parse URL scheme to determine checker type
    // Poll with exponential backoff: 100ms -> 200ms -> 400ms -> ... capped at 2s
    // All URLs must pass for success
}

// CheckHTTP - GET request, expect 2xx
func CheckHTTP(ctx context.Context, url string) error

// CheckGRPC - uses grpc.health.v1.Health/Check (can reuse executor/grpc.go code)
func CheckGRPC(ctx context.Context, url string) error

// CheckTCP - dial connection, immediate close on success
func CheckTCP(ctx context.Context, host string) error
```

### 3. Process Management

**New file:** `cli/internal/process/process.go`

```go
type ManagedProcess struct {
    cmd *exec.Cmd
}

func Start(ctx context.Context, command string) (*ManagedProcess, error) {
    // Spawn process with shell
    // Pipe stdout/stderr to parent (controlled by verbose flag)
    // Return handle for cleanup
}

func (p *ManagedProcess) Stop() error {
    // Send SIGTERM, wait, then SIGKILL if needed
}
```

### 4. Test Command Integration

**File:** `cli/cmd/yapi/test.go`

```go
func (a *app) runTest(cmd *cobra.Command, args []string) error {
    // Load project config
    project, _ := config.LoadProject(...)

    var process *process.ManagedProcess

    // If test.start configured and --no-start not set:
    if project.Test != nil && project.Test.Start != "" && !noStart {
        // 1. Start server
        process, err = process.Start(ctx, project.Test.Start)
        if err != nil {
            return fmt.Errorf("failed to start server: %w", err)
        }
        defer process.Stop()

        // 2. Wait for health
        timeout := parseTimeout(project.Test.Timeout, 60*time.Second)
        err = healthcheck.Wait(ctx, project.Test.WaitOn, timeout)
        if err != nil {
            return fmt.Errorf("health check failed: %w", err)
        }
    }

    // 3. Run tests (existing logic)
    return a.runTests(...)
}
```

### 5. New CLI Flags

**File:** `cli/internal/cli/commands/commands.go`

```go
testCmd.Flags().Bool("no-start", false, "Skip starting the dev server")
testCmd.Flags().String("start", "", "Command to start server (overrides config)")
testCmd.Flags().StringSlice("wait-on", nil, "URL(s) to wait for (overrides config)")
testCmd.Flags().Duration("wait-timeout", 60*time.Second, "Health check timeout")
```

## Files to Modify/Create

| File | Action |
|------|--------|
| `cli/internal/config/project.go` | Add TestConfig struct |
| `cli/internal/healthcheck/healthcheck.go` | New - health check polling |
| `cli/internal/process/process.go` | New - process management |
| `cli/cmd/yapi/test.go` | Integrate server startup |
| `cli/internal/cli/commands/commands.go` | Add new flags |

## Testing Strategy

1. **Unit tests:**
   - Config parsing with test block
   - Health check polling logic (mock HTTP server)
   - Process start/stop lifecycle

2. **Integration tests:**
   - Start a simple HTTP server, run test, verify cleanup
   - Timeout behavior when server never becomes healthy
   - `--no-start` flag behavior

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Field name | `start` | Matches GitHub Action, clear intent |
| Server output | Hidden, show with `--verbose` | Clean default, debug when needed |
| Multiple commands | Single string (use `&&` or scripts) | Simpler, shell handles complexity |
| Cleanup | Always kill server | Clean state, add `--keep-alive` later if needed |

## Documentation Updates

1. `README.md` - Add section on integrated test server
2. `SKILL.md` - Add examples for CI/CD and local dev workflows
