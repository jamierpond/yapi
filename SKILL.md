# yapi

CLI-first API testing for HTTP, GraphQL, gRPC, and TCP.

## The Workflow

yapi enables test-driven API development. Write the test first, then implement until it passes:

1. **Write the test** - Create a `.yapi.yml` file with the expected behavior
2. **Run it** - `yapi run file.yapi.yml` (it will fail)
3. **Implement/fix** - Build the API endpoint
4. **Iterate** - Refine assertions, add edge cases

This loop is the core of agentic API development with yapi.

---

## Environment Setup (Do This First)

Before writing any tests, set up your environments. Create `yapi.config.yml` in your project root:

```yaml
yapi: v1
default_environment: local

environments:
  local:
    url: http://localhost:3000
    vars:
      API_KEY: dev_key_123

  staging:
    url: https://staging.example.com
    vars:
      API_KEY: ${STAGING_API_KEY}  # from shell env

  prod:
    url: https://api.example.com
    vars:
      API_KEY: ${PROD_API_KEY}
    env_files:
      - .env.prod  # load secrets from file
```

Now your tests use `${url}` and `${API_KEY}` - same test, any environment:

```bash
yapi run get-users.yapi.yml              # uses local (default)
yapi run get-users.yapi.yml --env staging
yapi run get-users.yapi.yml --env prod
```

**Variable resolution order** (highest priority first):
1. Shell environment variables
2. Environment-specific `vars`
3. Environment-specific `env_files`
4. Default `vars`
5. Default `env_files`

---

## A) Smoke Testing

Quick health checks to verify endpoints are alive.

### HTTP

```yaml
yapi: v1
url: ${url}/health
method: GET
expect:
  status: 200
```

### GraphQL

```yaml
yapi: v1
url: ${url}/graphql
graphql: |
  query { __typename }
expect:
  status: 200
  assert:
    - .data.__typename != null
```

### gRPC

```yaml
yapi: v1
url: grpc://${host}:${port}
service: grpc.health.v1.Health
rpc: Check
plaintext: true
body:
  service: ""
expect:
  status: 200
```

### TCP

```yaml
yapi: v1
url: tcp://${host}:${port}
data: "PING\n"
encoding: text
expect:
  status: 200
```

---

## B) Integration Testing

Multi-step workflows with data passing between requests. Use chains when steps depend on each other.

### Authentication Flow

```yaml
yapi: v1
chain:
  - name: login
    url: ${url}/auth/login
    method: POST
    body:
      email: test@example.com
      password: ${TEST_PASSWORD}
    expect:
      status: 200
      assert:
        - .token != null

  - name: get_profile
    url: ${url}/users/me
    method: GET
    headers:
      Authorization: Bearer ${login.token}
    expect:
      status: 200
      assert:
        - .email == "test@example.com"
```

### CRUD Flow

```yaml
yapi: v1
chain:
  - name: create
    url: ${url}/posts
    method: POST
    body:
      title: "Test Post"
      content: "Hello World"
    expect:
      status: 201
      assert:
        - .id != null

  - name: read
    url: ${url}/posts/${create.id}
    method: GET
    expect:
      status: 200
      assert:
        - .title == "Test Post"

  - name: update
    url: ${url}/posts/${create.id}
    method: PATCH
    body:
      title: "Updated Post"
    expect:
      status: 200

  - name: delete
    url: ${url}/posts/${create.id}
    method: DELETE
    expect:
      status: 204
```

### Running Integration Tests

Name test files with `.test.yapi.yml` suffix:
```
tests/
  auth.test.yapi.yml
  posts.test.yapi.yml
  users.test.yapi.yml
```

Run all tests:
```bash
yapi test ./tests                    # sequential
yapi test ./tests --parallel 4       # concurrent
yapi test ./tests --env staging      # against staging
yapi test ./tests --verbose          # detailed output
```

---

## C) Uptime Monitoring

Create test suites for monitoring your services in production.

### Monitor Suite Structure

```
monitors/
  api-health.test.yapi.yml
  auth-service.test.yapi.yml
  database-check.test.yapi.yml
  graphql-schema.test.yapi.yml
```

### Health Check with Timeout

```yaml
yapi: v1
url: ${url}/health
method: GET
timeout: 5s  # fail if response takes longer
expect:
  status: 200
  assert:
    - .status == "healthy"
    - .database == "connected"
```

### Run Monitoring Suite

```bash
# Check all monitors in parallel
yapi test ./monitors --parallel 10 --env prod

# With verbose output for debugging
yapi test ./monitors --parallel 10 --env prod --verbose
```

### CI/CD Integration (GitHub Actions)

```yaml
name: API Health Check
on:
  schedule:
    - cron: '*/5 * * * *'  # every 5 minutes
  workflow_dispatch:

jobs:
  monitor:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install yapi
        run: curl -fsSL https://yapi.run/install/linux.sh | bash

      - name: Run health checks
        env:
          PROD_API_KEY: ${{ secrets.PROD_API_KEY }}
        run: yapi test ./monitors --env prod --parallel 5
```

### Load Testing

Stress test endpoints or entire workflows:

```bash
# 1000 requests, 50 concurrent
yapi stress api-flow.yapi.yml -n 1000 -p 50

# Run for 30 seconds
yapi stress api-flow.yapi.yml -d 30s -p 25

# Against production (with confirmation)
yapi stress api-flow.yapi.yml -e prod -n 500 -p 10
```

---

## D) Async Job Polling with `wait_for`

For endpoints that process data asynchronously, use `wait_for` to poll until conditions are met.

### Fixed Period Polling

```yaml
yapi: v1
url: ${url}/jobs/${job_id}
method: GET

wait_for:
  until:
    - .status == "completed" or .status == "failed"
  period: 2s
  timeout: 60s

expect:
  assert:
    - .status == "completed"
```

### Exponential Backoff

Better for rate-limited APIs or long-running jobs:

```yaml
yapi: v1
url: ${url}/jobs/${job_id}
method: GET

wait_for:
  until:
    - .status == "completed"
  backoff:
    seed: 1s       # Initial wait
    multiplier: 2  # 1s -> 2s -> 4s -> 8s...
  timeout: 300s
```

### Async Workflow Chain

Complete example: create job, poll until done, download result:

```yaml
yapi: v1
chain:
  - name: create_job
    url: ${url}/jobs
    method: POST
    body:
      type: "data_export"
      filters:
        date_range: "last_30_days"
    expect:
      status: 202
      assert:
        - .job_id != null

  - name: wait_for_job
    url: ${url}/jobs/${create_job.job_id}
    method: GET
    wait_for:
      until:
        - .status == "completed" or .status == "failed"
      period: 2s
      timeout: 300s
    expect:
      assert:
        - .status == "completed"
        - .download_url != null

  - name: download_result
    url: ${wait_for_job.download_url}
    method: GET
    output_file: ./export.csv
```

### Webhook/Callback Waiting

Wait for a webhook to be received:

```yaml
yapi: v1
chain:
  - name: trigger_action
    url: ${url}/payments/initiate
    method: POST
    body:
      amount: 100
    expect:
      status: 202

  - name: wait_for_webhook
    url: ${url}/webhooks/received
    method: GET
    wait_for:
      until:
        - . | length > 0
        - .[0].event == "payment.completed"
      period: 1s
      timeout: 30s
```

---

## E) Integrated Test Server

Automatically start your dev server, wait for health checks, run tests, and clean up. Configure in `yapi.config.yml`:

```yaml
yapi: v1

test:
  start: "npm run dev"
  wait_on:
    - "http://localhost:3000/healthz"
    - "grpc://localhost:50051"
  timeout: 60s
  parallel: 8
  directory: "./tests"

environments:
  local:
    url: http://localhost:3000
```

### Running with Integrated Server

```bash
# Automatically starts server, waits for health, runs tests, kills server
yapi test

# Skip server startup (server already running)
yapi test --no-start

# Override config from CLI
yapi test --start "npm start" --wait-on "http://localhost:4000/health"

# See server stdout/stderr
yapi test --verbose
```

### Health Check Protocols

| Protocol | URL Format | Behavior |
|----------|------------|----------|
| HTTP/HTTPS | `http://localhost:3000/healthz` | Poll until 2xx response |
| gRPC | `grpc://localhost:50051` | Uses `grpc.health.v1.Health/Check` |
| TCP | `tcp://localhost:5432` | Poll until connection succeeds |

### Local vs CI Parity

The same workflow works locally and in CI:

**Local development:**
```bash
yapi test  # starts server, runs tests, cleans up
```

**GitHub Actions:**
```yaml
- uses: jamierpond/yapi/action@main
  with:
    start: npm run dev
    wait-on: http://localhost:3000/healthz
    command: yapi test -a
```

---

## Commands Reference

| Command | Description |
|---------|-------------|
| `yapi run file.yapi.yml` | Execute a request |
| `yapi run file.yapi.yml --env prod` | Execute against specific environment |
| `yapi test ./dir` | Run all `*.test.yapi.yml` files |
| `yapi test ./dir --all` | Run all `*.yapi.yml` files (not just tests) |
| `yapi test ./dir --parallel 4` | Run tests concurrently |
| `yapi validate file.yapi.yml` | Check syntax without executing |
| `yapi watch file.yapi.yml` | Re-run on every file save |
| `yapi stress file.yapi.yml` | Load test with concurrency |
| `yapi list` | List all yapi files in directory |

---

## Assertion Syntax

Assertions use JQ expressions that must evaluate to true.

### Body Assertions

```yaml
expect:
  status: 200
  assert:
    - .id != null                    # field exists
    - .name == "John"                # exact match
    - .age > 18                      # comparison
    - . | length > 0                 # array not empty
    - .[0].email != null             # first item has email
    - .users | length == 10          # exactly 10 users
    - .type == "admin" or .type == "user"  # alternatives
    - .tags | contains(["api"])      # array contains value
```

### Header Assertions

```yaml
expect:
  status: 200
  assert:
    headers:
      - .["Content-Type"] | contains("application/json")
      - .["X-Request-Id"] != null
      - .["Cache-Control"] == "no-cache"
    body:
      - .data != null
```

### Status Code Options

```yaml
expect:
  status: 200           # exact match
  status: [200, 201]    # any of these
```

---

## Protocol Examples

### HTTP with Query Params and Headers

```yaml
yapi: v1
url: ${url}/api/users
method: GET
headers:
  Authorization: Bearer ${API_KEY}
  Accept: application/json
query:
  limit: "10"
  offset: "0"
  sort: "created_at"
expect:
  status: 200
```

### HTTP POST with JSON Body

```yaml
yapi: v1
url: ${url}/api/users
method: POST
body:
  name: "John Doe"
  email: "john@example.com"
  roles:
    - admin
    - user
expect:
  status: 201
  assert:
    - .id != null
```

### HTTP Form Data

```yaml
yapi: v1
url: ${url}/upload
method: POST
content_type: multipart/form-data
form:
  name: "document.pdf"
  description: "Q4 Report"
expect:
  status: 200
```

### GraphQL with Variables

```yaml
yapi: v1
url: ${url}/graphql
graphql: |
  query GetUser($id: ID!) {
    user(id: $id) {
      id
      name
      email
    }
  }
variables:
  id: "123"
expect:
  status: 200
  assert:
    - .data.user.id == "123"
```

### gRPC with Metadata

```yaml
yapi: v1
url: grpc://${host}:${port}
service: users.UserService
rpc: GetUser
plaintext: true
headers:
  authorization: Bearer ${API_KEY}
body:
  user_id: "123"
expect:
  status: 200
  assert:
    - .user.id == "123"
```

### TCP Raw Connection

```yaml
yapi: v1
url: tcp://${host}:${port}
data: |
  GET / HTTP/1.1
  Host: example.com

encoding: text
read_timeout: 5
expect:
  status: 200
```

---

## File Organization

Recommended project structure:

```
project/
  yapi.config.yml          # environments
  .env                     # local secrets (gitignored)
  .env.example             # template for secrets

  tests/
    auth/
      login.test.yapi.yml
      logout.test.yapi.yml
    users/
      create-user.test.yapi.yml
      get-user.test.yapi.yml

  monitors/
    health.test.yapi.yml
    critical-endpoints.test.yapi.yml
```

---

## Tips

- **Start simple**: Begin with status code checks, add body assertions as needed
- **Use watch mode**: `yapi watch file.yapi.yml` for rapid iteration
- **Validate before running**: `yapi validate file.yapi.yml` catches syntax errors
- **Keep tests focused**: One logical flow per file
- **Name steps clearly**: In chains, use descriptive names like `create_user`, `verify_email`
- **Reference previous steps**: Use `${step_name.field}` to pass data between chain steps
