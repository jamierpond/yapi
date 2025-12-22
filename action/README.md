# Yapi GitHub Action

GitHub Action to run [Yapi](https://yapi.run) integration tests with automatic service orchestration and health checks.

## Features

- Automatically installs the Yapi CLI
- Starts background services (Node servers, Python backends, etc.)
- Waits for health check URLs before running tests
- Executes Yapi test suites with full output
- Fails the workflow if tests fail

## Usage

### Basic Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Yapi Tests
        uses: jamierpond/yapi-action@v1
        with:
          command: yapi run ./tests
```

### With Background Services

```yaml
- uses: jamierpond/yapi-action@v1
  with:
    start: |
      npm run dev
      python api/server.py
    wait-on: |
      http://localhost:3000
      http://localhost:8000/health
    command: yapi run ./tests/integration
```

### Full Example with Multiple Services

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install dependencies
        run: npm install

      - name: Run Integration Tests
        uses: jamierpond/yapi-action@v1
        with:
          start: |
            pnpm --filter web dev
            cd media-service && uv run main.py
          wait-on: |
            http://localhost:3000
            http://localhost:8000/healthz
          wait-on-timeout: 120000
          command: yapi run ./tests/integration
```

## Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `start` | No | `""` | Multiline string of commands to start in background (e.g., `npm run dev`) |
| `wait-on` | No | `""` | Multiline string of URLs to poll before testing (e.g., `http://localhost:3000`) |
| `wait-on-timeout` | No | `60000` | Timeout in milliseconds for wait-on health checks |
| `command` | No | `yapi run .` | The main test command to execute |
| `version` | No | `latest` | Version of yapi to install |

## How It Works

1. **Install Yapi**: Downloads and installs the Yapi CLI binary
2. **Start Services**: Spawns background processes (servers, databases, etc.) in parallel
3. **Wait for Health**: Polls specified URLs until they return 200 OK or timeout
4. **Run Tests**: Executes your Yapi test suite
5. **Report Results**: Fails the workflow if tests fail

## Examples

### Simple API Tests

```yaml
- uses: jamierpond/yapi-action@v1
  with:
    command: yapi run api-tests.yapi.yml
```

### With Local Server

```yaml
- uses: jamierpond/yapi-action@v1
  with:
    start: npm start
    wait-on: http://localhost:3000
    command: yapi run tests/
```

### Multiple Services with Custom Timeout

```yaml
- uses: jamierpond/yapi-action@v1
  with:
    start: |
      docker-compose up -d
      npm run backend
    wait-on: |
      http://localhost:8080/health
      http://localhost:5432
    wait-on-timeout: 180000
    command: yapi run integration/
```

## Development

To work on this action locally:

```bash
# Install dependencies
pnpm install

# Build the action
pnpm run build

# The compiled output will be in dist/
```

## License

MIT
