# URGENT: DRY out config execution logic

## The Bug

`yapi watch` (pretty TUI mode) was sending `${GH_TOKEN}` literally instead of substituting the environment variable value. This caused 401 "Bad credentials" errors when running from the Neovim plugin, while `yapi run` worked fine.

## Root Cause

There are **two separate implementations** of config execution:

1. `cmd/yapi/main.go` - `runConfigPath()` and `runConfigPathSafe()`
2. `internal/tui/watch.go` - `runYapiCmd()` and `executeConfig()`

The TUI version was missing `cfg.SubstituteEnvVars()` before execution.

## The Fix Applied

Added `cfg.SubstituteEnvVars()` to `watch.go:107`.

## Required Refactor

This code duplication is dangerous. Both files have:
- Config loading
- Validation
- Env var substitution
- Transport detection
- Executor dispatch
- JQ filtering
- Output highlighting

**There should be ONE function that handles execution.** Something like:

```go
// internal/runner/runner.go
func Execute(cfg *config.YapiConfig, opts ExecuteOpts) (string, error) {
    // validate
    // substitute env vars
    // detect transport
    // execute
    // apply jq filter
    // highlight output
}
```

Both `main.go` and `watch.go` should call this single function.

## Files to consolidate

- `cli/cmd/yapi/main.go` lines 156-207, 219-268, 350-409
- `cli/internal/tui/watch.go` lines 88-148

## Risk if not addressed

Any future changes to execution logic (new protocols, new preprocessing steps, new options) will need to be applied in multiple places, and one will inevitably be missed.
