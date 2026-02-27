# yapi — LLM Skill Guide

yapi is a CLI-first, git-friendly API client. You define requests in YAML files and run them from the terminal. No GUI, no accounts, no state — just files and a binary.

## When to Use yapi

- Send HTTP, gRPC, GraphQL, or TCP requests
- Chain multiple requests together (auth flow, then use the token)
- Assert on responses (status codes, body content via JQ)
- Run API test suites with `yapi test`
- Poll endpoints until a condition is met

## Core Concepts

**Every request file starts with `yapi: v1`.** This is required.

**File extension:** `.yapi.yml` (or `.yapi.yaml`). Test files use `.test.yapi.yml`.

**Project config:** `yapi.config.yml` at the project root defines environments and base URLs. yapi walks up the directory tree to find it.

**Variables:** Use `${VAR}` syntax. Resolved from: chain step outputs > environment vars > shell env > defaults. Default values: `${VAR:-fallback}`.

## Quick Reference

### Run a request file

```bash
yapi run request.yapi.yml
yapi run request.yapi.yml -e prod   # with environment
```

### Quick one-off request (no file needed)

```bash
yapi send https://api.example.com/users           # GET
yapi send https://api.example.com/users '{"name":"Alice"}'  # POST (auto-detected)
yapi send -X PUT https://api.example.com/users/1 '{"name":"Bob"}' -H 'Authorization: Bearer tok'
yapi send https://api.example.com/users --jq '.[0].name'
```

### Minimal request file

```yaml
yapi: v1
url: https://api.example.com/users
method: GET
```

### POST with body

```yaml
yapi: v1
url: https://api.example.com/users
method: POST
body:
  name: Alice
  role: admin
```

### Chain requests (pass data between steps)

```yaml
yapi: v1
chain:
  - name: login
    url: https://api.example.com/auth
    method: POST
    body:
      username: ${USERNAME}
      password: ${PASSWORD}
    expect:
      status: 200

  - name: get_profile
    url: https://api.example.com/me
    method: GET
    headers:
      Authorization: Bearer ${login.token}
    expect:
      status: 200
      assert:
        - .email != null
```

### Assert on responses

```yaml
expect:
  status: 200
  assert:
    - . | length > 0        # JQ expression, must evaluate to true
    - .[0].id != null
    - .count >= 10
```

### Run tests

```bash
yapi test ./tests              # run all *.test.yapi.yml files
yapi test ./tests -e staging   # against a specific environment
yapi test ./tests -p 8         # parallel execution
```

### Environments

Define in `yapi.config.yml`:

```yaml
yapi: v1
default_environment: local
environments:
  local:
    url: http://localhost:3000
    vars:
      API_KEY: dev_key
  prod:
    url: https://api.example.com
    vars:
      API_KEY: ${PROD_API_KEY}
```

Then in request files, use `${url}` and `${API_KEY}` — they resolve from the active environment.

## Documentation Map

The docs are split into **topics** (concepts/features) and **commands** (CLI reference). Read only what you need.

### Topics (`docs/topics/`)

| File | Read this when you need to... |
|------|-------------------------------|
| `config.md` | Know all available YAML fields (full schema reference) |
| `chain.md` | Chain multiple requests, pass data between steps |
| `assert.md` | Write assertions on status, body, or headers |
| `variables.md` | Understand `${VAR}` interpolation and resolution order |
| `environments.md` | Set up dev/staging/prod environments |
| `jq.md` | Filter or transform response bodies with JQ |
| `testing.md` | Use the built-in test runner (`yapi test`) |
| `polling.md` | Poll an endpoint until a condition is met (`wait_for`) |
| `send.md` | Use `yapi send` for quick curl-like requests |
| `protocols.md` | Use gRPC, GraphQL, or TCP (not just HTTP) |

### Commands (`docs/commands/`)

Each file documents one CLI command: `yapi_run.md`, `yapi_send.md`, `yapi_test.md`, `yapi_stress.md`, `yapi_watch.md`, etc. Refer to these for flag details and usage examples.

## Common Patterns

**Auth flow then use token:**
Chain with `${login.token}` in the Authorization header of subsequent steps.

**Validate an API contract:**
Use `expect.assert` with JQ expressions. Chain variable assertions let you compare across steps: `.id == ${create.id}`.

**Wait for async job:**
Use `wait_for.until` with a `period` or `backoff` and `timeout`.

**Filter noisy responses:**
Use `jq_filter` in the YAML or `--jq` on the CLI.

## Gotchas

- `yapi: v1` at the top of every file — missing it is the most common error
- Variables are `${VAR}`, not `$VAR`
- Assertions are JQ expressions that must evaluate to `true`
- Chains stop on first failure (fail-fast)
- Protocol is auto-detected from URL scheme: `grpc://`, `tcp://`, or HTTP by default
