# Testing

yapi has a built-in test runner for running assertion-based API tests,
with parallel execution and dev server management.

## Test Files

By default, `yapi test` runs files matching `*.test.yapi`, `*.test.yapi.yml`,
or `*.test.yapi.yaml`:

```bash
yapi test              # Current directory
yapi test ./tests      # Specific directory
yapi test -a           # All .yapi files, not just .test ones
```

## Writing Tests

A test file is a normal yapi request file with `expect` assertions:

```yaml
# users.test.yapi.yml
yapi: v1
url: https://api.example.com/users
method: GET
expect:
  status: 200
  assert:
    - . | type == "array"
    - . | length > 0
```

Chain files work as tests too — each step's assertions are checked:

```yaml
# auth-flow.test.yapi.yml
yapi: v1
chain:
  - name: login
    url: /auth/login
    method: POST
    body:
      email: ${TEST_EMAIL}
      password: ${TEST_PASSWORD}
    expect:
      status: 200
      assert:
        - .token != null

  - name: me
    url: /auth/me
    headers:
      Authorization: Bearer ${login.token}
    expect:
      status: 200
```

## Parallel Execution

Run tests concurrently:

```bash
yapi test -p 4         # 4 parallel threads
```

## Dev Server Management

yapi can start your dev server and wait for it before running tests.

### Via yapi.config.yml

```yaml
# yapi.config.yml
yapi: v1
default_environment: local

test:
  start: npm run dev
  wait_on:
    - http://localhost:3000/health
  wait_timeout: 30s
```

### Via CLI Flags

```bash
yapi test --start "npm run dev" --wait-on http://localhost:3000/health
yapi test --no-start             # Skip server startup
```

The server is automatically stopped when tests finish or on Ctrl+C.

## Verbose Output

```bash
yapi test -v           # Show detailed pass/fail per test
```

## Environment Selection

```bash
yapi test -e staging   # Run tests against staging environment
```

## GitHub Actions

```yaml
- uses: jamierpond/yapi/action@0.X.X
  with:
    start: npm run dev
    wait-on: http://localhost:3000/health
    command: yapi test ./tests -a
```

## Exit Codes

- `0`: All tests passed
- `1`: One or more tests failed

## See Also

- `yapi docs assert` — Assertions on status, body, and headers
- `yapi docs chain` — Multi-step request chaining
- `yapi docs environments` — Running tests across environments
