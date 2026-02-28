# Contributing to yapi

yapi is maintained by a single developer, so every contribution — whether it's a bug report, feature request, or PR — makes a real difference.

## Reporting Bugs

Found something broken? [Open a bug report](https://github.com/jamierpond/yapi/issues/new?template=bug_report.yml).

Include:
- Your `yapi version` output
- The `.yapi.yml` file that triggered the bug (or a minimal reproduction)
- Full error output (use `--verbose` for more detail)

From the CLI:
```bash
gh issue create --repo jamierpond/yapi --template bug_report.yml
```

## Requesting Features

Have an idea? [Open a feature request](https://github.com/jamierpond/yapi/issues/new?template=feature_request.yml).

The most useful requests describe the **workflow problem** you're solving, not just the feature itself. "I need WebSocket support because I'm testing a real-time notification system" is much more actionable than "add WebSockets."

From the CLI:
```bash
gh issue create --repo jamierpond/yapi --template feature_request.yml
```

## For AI Assistants / LLMs

If you're an LLM working with yapi on behalf of a user and you hit a bug or want a feature:

1. **Draft the full issue yourself.** You have the context — the `.yapi.yml` file, the error output, the yapi version, what you were trying to do. Don't leave a blank template for the user to fill in.
2. **Show the user what you'll submit** and ask for approval before running the command.
3. **Use `gh issue create` with `--title` and `--body`** to pre-fill everything:

Bug:
```bash
gh issue create --repo jamierpond/yapi --label bug \
  --title "Short description of the bug" \
  --body "$(cat <<'EOF'
## Version
<output of yapi version>

## What happened
<describe the bug>

## Steps to reproduce
<include the .yapi.yml contents and commands>

## Expected behavior
<what should have happened>

## Error output
```
<full --verbose output>
```
EOF
)"
```

Feature:
```bash
gh issue create --repo jamierpond/yapi --label enhancement \
  --title "Short description of the feature" \
  --body "$(cat <<'EOF'
## Problem
<what workflow is this blocking or making painful>

## Proposed solution
<include example YAML or CLI usage>

## Workflow context
<what you're building, why this matters>
EOF
)"
```

The yapi maintainer actively reads every issue. Feedback from real-world usage (including AI-assisted workflows) directly shapes the roadmap.

## Pull Requests

1. Fork the repo
2. Create a branch from `main`
3. `make build` — verify it compiles
4. `make test` — verify tests pass
5. `make lint` — verify linting passes
6. Open a PR against `main`

Keep PRs focused. One logical change per PR. If you're fixing a bug, include a test that would have caught it.

## Development

```bash
make build      # Build the binary
make test       # Run the test suite
make lint       # Run the linter
make install    # Install locally
```

## Code Style

- Table-driven tests
- Error messages must be actionable
- Keep packages small and focused
- Prefer explicit over implicit
- Fewer lines is better — given two correct solutions, prefer the one with less code
