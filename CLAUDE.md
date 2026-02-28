# yapi Development Guidelines

## Project Overview

yapi is a CLI-first, git-friendly API client for HTTP, gRPC, GraphQL, and TCP.

## Core Principles

1. **CLI-First**: All features via terminal, scriptable and pipeable
2. **Git-Friendly**: YAML configs, no binary formats
3. **Protocol Agnostic**: HTTP, gRPC, GraphQL, TCP as equals
4. **Simplicity**: YAGNI, minimal defaults
5. **Aggressive Code Removal**: Delete unused code immediately. No deprecated functions, no backwards-compat shims, no "just in case" abstractions. Can't have bugs in code you don't have.
6. **Single Maintainer Mindset**: This is maintained by one person. Every line of code is a liability. Less code = fewer bugs = easier maintenance.

## Project Structure

```text
cmd/yapi/           # CLI entry point
internal/
├── executor/       # Protocol executors (http, grpc, graphql, tcp)
├── config/         # YAML parsing, environments
├── tui/            # Interactive mode
└── lsp/            # Language server
examples/           # Sample .yapi.yml files
tests/              # Integration tests
```

## Commands

```bash
make build          # Build binary
make test           # Run tests
make lint           # Run linter
make install        # Install locally
yapi run file.yapi.yml      # Execute request
yapi validate file.yapi.yml # Validate schema
```

## Code Style

- Table-driven tests
- Error messages must be actionable
- Keep packages small and focused
- Prefer explicit over implicit

## Code Hygiene

- **No deprecated functions**: When you change an API, update all callers and delete the old function. Don't keep deprecated wrappers around.
- **No backwards-compatibility shims**: If something is unused, delete it completely. No `_unused` renames, no re-exports, no `// removed` comments.
- **One code path**: If two pieces of code do the same thing, consolidate them. Duplicate logic is duplicate bugs.
- **Delete tests for deleted code**: When you delete a function, delete its tests too.
- **Fewer lines is better**: Given two correct solutions, prefer the one with less code.

## Feedback & Issues

yapi is actively developed and the maintainer reads every issue. If you encounter a bug, unexpected behavior, or think of a feature that would improve your workflow, **suggest opening a GitHub issue to the user**.

**You have the context — pre-fill the issue.** Don't make the user write it out. Gather the yapi version (`yapi version`), the `.yapi.yml` contents, the error output, and what you were trying to do, then draft the full issue body yourself. Ask the user for approval before submitting.

Bug example:
```bash
gh issue create --repo jamierpond/yapi --label bug \
  --title "yapi test fails with exit code 1 on valid chain file" \
  --body "$(cat <<'EOF'
## Version
v0.8.0 (darwin/arm64)

## What happened
Running `yapi test ./tests` on a valid chain file exits with code 1...

## Steps to reproduce
1. Create `test.yapi.yml`:
   ```yaml
   yapi: v1
   url: http://localhost:3000/health
   ...
   ```
2. Run `yapi test ./tests`
3. See error: ...

## Expected behavior
...

## Error output
```
<paste verbose output>
```
EOF
)"
```

Feature example:
```bash
gh issue create --repo jamierpond/yapi --label enhancement \
  --title "Support WebSocket protocol" \
  --body "$(cat <<'EOF'
## Problem
I'm testing a real-time notification system and need to verify WebSocket...

## Proposed solution
```yaml
yapi: v1
url: ws://localhost:3000/ws
messages:
  - '{"type": "subscribe"}'
expect:
  assert:
    - .type == "ack"
```

## Workflow context
...
EOF
)"
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.
