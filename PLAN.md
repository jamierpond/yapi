# Plan: Better Error Messages & Debuggability

## Context
Implementing WISHLIST.md items #1, #2, and #4 (removing #3 since `yapi send` already exists).

---

## 1. WISHLIST.md cleanup
Remove item #3 (yapi send) since it's already shipped.

---

## 2. Warn on bare `$word.word` variable syntax (WISHLIST #1)

The problem: `$step.field` (no braces) silently passes as a literal string instead of being substituted. Only `${step.field}` works.

**Changes:**

- **`cli/internal/vars/vars.go`**: Add a `BareChainRef` regex that matches `$word.word` patterns that are NOT inside `${...}`.
- **`cli/internal/vars/vars.go`**: Add `FindBareRefs(s string) []string` that returns the bare refs found.
- **`cli/internal/validation/analyzer.go`**: In `analyzeParsed()`, call a new `warnBareChainRefs(text)` validation function that scans the raw YAML text for bare `$word.word` patterns and emits `SeverityWarning` diagnostics with line numbers and an actionable message like:
  `"possible bare variable reference '$step.field' -- did you mean '${step.field}'? Only the ${...} form is substituted."`

This catches the problem at config analysis time (before execution), so users see the warning immediately -- even in `yapi validate`.

---

## 3. Show resolved request details in verbose chain execution (WISHLIST #2)

The problem: When a chain step fails, you can't see what values were actually sent because variable substitution is invisible.

**Changes:**

- **`cli/internal/runner/runner.go`**: Add `Verbose bool` field to `runner.Options`.
- **`cli/internal/runner/runner.go`**: In `RunChain()`, after `interpolateConfig()` succeeds and before executing, if `opts.Verbose` is true, print the resolved config to stderr:
  - Resolved URL (with method)
  - Resolved headers
  - Resolved body (JSON-serialized if map, or raw if string)
  - Uses `fmt.Fprintf(os.Stderr, ...)` with `[VERBOSE]` prefix, consistent with the existing Logger pattern.
- **`cli/cmd/yapi/run.go`**: Set `opts.Verbose = ctx.verbose` when building runner.Options in `executeRunE()`.

---

## 4. Print step responses in verbose chain mode (WISHLIST #4)

The problem: In chain execution, you only see the final failing step's output, not intermediate step responses.

**Changes:**

- **`cli/internal/runner/runner.go`**: In `RunChain()`, after each step executes, if `opts.Verbose` is true, print the step's response details to stderr:
  - Status code
  - Response body (truncated at 1000 chars for readability)
  - Duration

  This replaces the need for a per-step `debug: true` field -- verbose mode shows everything, which is simpler and avoids new config surface area.

---

## 5. Example files: `examples/debugging/`

Create example `.yapi.yml` files that demonstrate the improved debugging experience:

- **`bare-variable-warning.yapi.yml`**: A chain that uses `$step.field` (bare) to trigger the new warning.
- **`chain-verbose-demo.yapi.yml`**: A multi-step chain against jsonplaceholder with variables that shows how `--verbose` reveals resolved values.
- **`assertion-failure-demo.yapi.yml`**: A request with an `expect:` block that will fail, showing the detailed assertion error output.
- **`missing-key-demo.yapi.yml`**: A chain that references a nonexistent JSON key, showing the precise error path.

---

## Files changed

| File | Change |
|------|--------|
| `WISHLIST.md` | Remove item #3 |
| `cli/internal/vars/vars.go` | Add `BareChainRef` regex + `FindBareRefs()` function |
| `cli/internal/validation/analyzer.go` | Add `warnBareChainRefs()`, call from `analyzeParsed()` |
| `cli/internal/runner/runner.go` | Add `Verbose` to `Options`, add verbose logging in `RunChain()` |
| `cli/cmd/yapi/run.go` | Thread `verbose` into `runner.Options` |
| `examples/debugging/*.yapi.yml` | 4 new example files |

**No new dependencies. No config schema changes. No breaking changes.**
