# Yapi GitHub Action

Run [Yapi](https://yapi.run) integration tests in GitHub Actions with automatic service orchestration and health checks.

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

### With Background Service

```yaml
- uses: jamierpond/yapi/action@v0.5.0
  with:
    start: npm run dev
    wait-on: http://localhost:3000/health
    command: yapi test ./tests
```

### Multiple Services

```yaml
- uses: jamierpond/yapi/action@v0.5.0
  with:
    start: |
      npm run api
      python worker.py
    wait-on: |
      http://localhost:8080/health
      http://localhost:9000/ready
    command: yapi test ./integration
```

## Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `command` | No | `yapi test .` | The yapi command to run |
| `start` | No | `""` | Commands to run in background (one per line) |
| `wait-on` | No | `""` | URLs to wait for before running tests (one per line) |
| `wait-on-timeout` | No | `60000` | Health check timeout in milliseconds |

## How It Works

1. Installs yapi CLI
2. Starts background services
3. Waits for health checks
4. Runs your tests
5. Fails if tests fail

