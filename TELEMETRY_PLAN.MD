This is a great instinct. The current implementation in the provided logs and diffs is **invasive**. It pollutes `main.go`, `ConfigV1`, and the `Runner` with checks like `if app.tracker != nil`. It creates "temporal coupling" (you must call Start before End) and mixes business logic with instrumentation.

Since we are in `0.0.X`, we should absolutely be radical. We can achieve a "self-instrumenting" architecture using the **Middleware/Decorator pattern** and **Context-aware telemetry**.

Here is the plan to re-architect this into a clean, invisible, and robust system.

### The Architecture Plan

1.  **The "Telemetry" Package (Global & Safe)**

      * Rename `internal/analytics` to `internal/telemetry`.
      * It should be fail-safe. If keys aren't present, it does nothing (no-op).
      * It should support a `Context` driven workflow.

2.  **CLI Middleware (The "Onion" Layer)**

      * Instead of manual calls in `PersistentPreRun`/`PostRun`, we wrap the Cobra command runners.
      * This automatically captures arguments, flags, execution time, and **errors** for *every* command without touching `main.go` logic.

3.  **Decoupling Config from Stats**

      * Remove `CollectStats()` from `ConfigV1`. Configuration data structs should not know about analytics.
      * Create a `telemetry.Analyze(config)` function that inspects a config and returns properties.

4.  **LSP Interceptor**

      * Wrap the LSP `protocol.Handler` to automatically track session starts, document saves, and validation errors.

-----

### Step 1: The Telemetry Package

We need a cleaner interface. We will use a `context` to store the "Current Trace" so we don't have to pass tracker objects around manually.

**`internal/telemetry/telemetry.go`**

```go
package telemetry

import (
	"context"
	"os"
	"time"
    // ... imports
)

var client posthog.Client

// Init initializes the global client (call in main.go)
func Init(version, commit string) {
    if os.Getenv("YAPI_NO_ANALYTICS") != "" { return }
    // ... logic to setup PostHog with Build Tags ...
}

// Close ensures all events are flushed
func Close() {
    if client != nil { client.Close() }
}

// Track emits an event fire-and-forget style
func Track(event string, props map[string]interface{}) {
    if client == nil { return }
    // ... enrich with OS, Arch, Version ...
    client.Enqueue(posthog.Capture{Event: event, Properties: props})
}

// --- Context Awareness ---

type contextKey struct{}
type Trace struct {
    Command   string
    StartTime time.Time
    Props     map[string]interface{}
}

// WithContext creates a new telemetry scope
func WithContext(ctx context.Context, commandName string, props map[string]interface{}) context.Context {
    return context.WithValue(ctx, contextKey{}, &Trace{
        Command:   commandName,
        StartTime: time.Now(),
        Props:     props,
    })
}

// EndFromContext retrieves the trace, calculates duration, and sends the event
func EndFromContext(ctx context.Context, err error) {
    tr, ok := ctx.Value(contextKey{}).(*Trace)
    if !ok { return }

    duration := time.Since(tr.StartTime)
    tr.Props["duration_ms"] = duration.Milliseconds()
    tr.Props["success"] = (err == nil)
    if err != nil {
        tr.Props["error"] = err.Error()
    }

    Track(tr.Command, tr.Props)
}
```

### Step 2: The Cobra Middleware (CLI)

This removes 90% of the code from your `main.go`. We create a wrapper that inspects the Cobra command and injects telemetry.

**`internal/cli/middleware/telemetry.go`**

```go
package middleware

import (
    "github.com/spf13/cobra"
    "yapi.run/cli/internal/telemetry"
)

// WrapWithTelemetry automatically adds timing and error reporting to a command
func WrapWithTelemetry(cmd *cobra.Command) {
    // Wrap children recursively (handles "yapi run", "yapi watch", etc)
    for _, c := range cmd.Commands() {
        WrapWithTelemetry(c)
    }

    // Don't instrument help or completion commands
    if cmd.Name() == "help" || cmd.Name() == "completion" {
        return
    }

    // Save the original RunE (standardize on RunE for error handling)
    originalRun := cmd.RunE
    // If command uses Run instead of RunE, convert it
    if originalRun == nil && cmd.Run != nil {
        originalRun = func(c *cobra.Command, args []string) error {
            cmd.Run(c, args)
            return nil
        }
    }

    if originalRun != nil {
        cmd.RunE = func(c *cobra.Command, args []string) error {
            // 1. Gather Flags
            props := make(map[string]interface{})
            cmd.Flags().Visit(func(f *pflag.Flag) {
                props["flag_"+f.Name] = f.Value.String()
            })
            props["args_count"] = len(args)

            // 2. Start Context
            ctx := telemetry.WithContext(c.Context(), "cmd_"+cmd.Name(), props)
            c.SetContext(ctx)

            // 3. Run Command
            err := originalRun(c, args)

            // 4. Auto-End (captures error and duration)
            telemetry.EndFromContext(ctx, err)

            return err
        }
    }
}
```

**New `main.go`:**

```go
func main() {
    telemetry.Init(version, commit)
    defer telemetry.Close()

    rootCmd := &cobra.Command{Use: "yapi", ...}
    // ... add commands ...

    // ONE LINE to rule them all:
    middleware.WrapWithTelemetry(rootCmd)

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Step 3: Logic Instrumentation (The Engine)

Instead of `main.go` knowing about request stats, let the `Engine` record what it does. The `Engine` is already doing the work, so it knows best.

**`internal/core/engine.go`**

```go
func (e *Engine) RunConfig(ctx context.Context, path string, opts runner.Options) *RunConfigResult {
    // Parse
    analysis, err := validation.AnalyzeConfigFile(path)

    // EXTRACT STATS HERE (Decoupled from ConfigV1)
    // We do this silently in the background
    stats := telemetry.ExtractConfigStats(analysis)

    // Create a specific trace for the execution phase
    // Note: We use the ctx passed from Cobra, so this is a "child" span conceptually
    defer func(start time.Time) {
         stats["duration_ms"] = time.Since(start).Milliseconds()
         stats["success"] = (result.Error == nil)
         telemetry.Track("execution_run", stats)
    }(time.Now())

    // ... existing execution logic ...
}
```

**`internal/telemetry/analyzer.go`** (New File)
This file inspects `ConfigV1` and generates a map. It keeps `ConfigV1` clean.

```go
func ExtractConfigStats(a *validation.Analysis) map[string]interface{} {
    m := make(map[string]interface{})
    if a.Request != nil {
        m["transport"] = a.Request.Metadata["transport"]
    }
    m["is_chain"] = len(a.Chain) > 0
    // ... logic moved from CollectStats ...
    return m
}
```

### Step 4: LSP Integration

Since LSP runs as a long-lived process, we hook into the `handler`.

**`internal/langserver/langserver.go`**

```go
func Run() {
    // ... setup ...
    handler = protocol.Handler{
        Initialize: wrapHandler("initialize", initialize),
        TextDocumentDidSave: wrapHandler("did_save", textDocumentDidSave),
        // ...
    }
}

// Generic wrapper for LSP handlers
func wrapHandler(name string, fn func(ctx *glsp.Context, params interface{}) (interface{}, error)) ... {
    return func(ctx, params) {
        start := time.Now()
        res, err := fn(ctx, params)
        telemetry.Track("lsp_"+name, map[string]interface{}{
            "duration_ms": time.Since(start).Milliseconds(),
            "error": err != nil,
        })
        return res, err
    }
}
```

### Benefits of this approach

1.  **Invisible:** `main.go` doesn't know about `requestTracker`, `status`, or `errors`. It just runs commands.
2.  **Complete:** Because we wrap the Cobra `RunE`, *every* command (even new ones you add later) is automatically instrumented with timing and error capturing.
3.  **DRY:** Stats extraction logic lives in `telemetry`, not scattered across `Config` and `main`.
4.  **Resilient:** If `telemetry.Init` isn't called (e.g., in tests), `Track` is a no-op. No nil pointer dereferences.

### Next Step

Would you like me to implement the **Cobra Middleware** (`internal/cli/middleware/telemetry.go`) and the **Global Telemetry Package** first so we can strip the noisy code out of `main.go`?
