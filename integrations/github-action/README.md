# Yapi GitHub Action

Run [Yapi](https://yapi.run) integration tests in GitHub Actions with automatic service orchestration.

## Usage

**Use the latest version from main:**

```yaml
- uses: jamierpond/yapi/action@main
  with:
    command: yapi test ./tests
```

**Or use a specific version tag:**

```yaml
- uses: jamierpond/yapi/action@v0.5.0
  with:
    command: yapi test ./tests
```

The action automatically installs the matching yapi version based on the action ref you use. For example, `@v0.5.0` installs yapi v0.5.0, while `@main` installs the latest version.

### Basic Example

```yaml
name: Integration Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: jamierpond/yapi/action@v0.5.0
        with:
          command: yapi test ./tests
```

### With Background Service and Health Checks

Use yapi's native `--wait-on` flag for health checks:

```yaml
- uses: jamierpond/yapi/action@v0.5.0
  with:
    start: npm run dev
    command: yapi test ./tests --wait-on=http://localhost:3000/health
```

### Multiple Services

```yaml
- uses: jamierpond/yapi/action@v0.5.0
  with:
    start: |
      npm run api
      python worker.py
    command: yapi test ./integration --wait-on=http://localhost:8080/health --wait-on=http://localhost:9000/ready
```

### With Custom Timeout

```yaml
- uses: jamierpond/yapi/action@v0.5.0
  with:
    start: npm run dev
    command: yapi test ./tests --wait-on=http://localhost:3000/health --wait-timeout=120s
```

## Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `command` | No | `yapi test .` | The yapi command to run. Use `--wait-on` for health checks. |
| `start` | No | `""` | Commands to run in background (one per line) |
| `skip-install` | No | `false` | Skip yapi installation (use pre-installed version) |

## How It Works

1. Installs yapi CLI (or uses pre-installed version)
2. Starts background services
3. Runs your yapi command (which handles health checks via `--wait-on`)
4. Fails if tests fail

## Health Check Options

Yapi's `test` command supports native health checks:

- `--wait-on=URL` - Wait for URL(s) to be ready (http://, grpc://, tcp://)
- `--wait-timeout=DURATION` - Health check timeout (default: 60s)

```bash
yapi test ./tests --wait-on=http://localhost:3000/healthz --wait-timeout=90s
```
