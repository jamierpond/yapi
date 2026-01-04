# yapi Developer Guide

## What is yapi?

**yapi** is a CLI-first, git-friendly API client that uses YAML files to define and execute HTTP, gRPC, GraphQL, and TCP requests. Unlike traditional API clients (Postman, Insomnia), yapi stores all requests as version-controllable YAML files, making it ideal for:

- API testing in CI/CD pipelines
- Integration test suites
- API documentation as code
- Multi-environment request management
- Request chaining and workflow automation

## Core Concepts

### 1. Request Files (`.yapi.yml` / `.yapi.yaml`)

Every yapi request is defined in a YAML file with the `.yapi.yml` or `.yapi.yaml` extension.

**Mandatory requirement:** All files MUST start with `yapi: v1` to enable schema evolution.

### 2. Basic Request Structure

```yaml
yapi: v1
url: https://api.example.com/endpoint
method: GET  # GET, POST, PUT, PATCH, DELETE
timeout: 4s  # Optional: Request timeout (e.g., "4s", "100ms", "1m")
headers:
  Authorization: Bearer ${TOKEN}
body:
  key: value
expect:
  status: 200
  assert:
    - .data != null
```

### 3. Environment Configuration (`yapi.config.yml`)

Projects should have a `yapi.config.yml` file in the root directory to manage environments:

```yaml
yapi: v1

default_environment: local

defaults:
  vars:
    SHARED_VAR: shared_value

environments:
  local:
    url: http://localhost:3000
    vars:
      API_KEY: local_key

  prod:
    url: https://api.example.com
    vars:
      API_KEY: ${PROD_API_KEY}  # From shell environment
    env_file: .env.prod
```

**Key points:**
- `default_environment`: Which environment to use when `-e` flag is not specified
- `defaults`: Variables available in ALL environments
- `environments`: Environment-specific settings
- Variables can reference shell environment vars with `${VAR_NAME}`
- Default values: `${VAR_NAME:-default_value}`

## Variable Interpolation

Variables are interpolated using `${VAR_NAME}` syntax:

```yaml
yapi: v1
url: ${url}/api/users/${USER_ID}
headers:
  Authorization: Bearer ${API_KEY}
  X-Custom: ${MY_VAR:-default}
```

**Variable sources (in priority order):**
1. Chain step references: `${step_name.field}`
2. Environment vars from `yapi.config.yml`
3. Shell environment variables
4. Default values specified with `:-`

## Request Chaining

One of yapi's most powerful features is the ability to chain requests and pass data between them:

```yaml
yapi: v1
chain:
  - name: login
    url: https://api.example.com/auth/login
    method: POST
    body:
      username: ${USERNAME}
      password: ${PASSWORD}
    expect:
      status: 200
      assert:
        - .token != null

  - name: create_post
    url: https://api.example.com/posts
    method: POST
    headers:
      Authorization: Bearer ${login.token}  # Reference from login step
    body:
      title: "My Post"
    expect:
      status: 201
      assert:
        - .id != null

  - name: verify_post
    url: https://api.example.com/posts/${create_post.id}
    method: GET
    headers:
      Authorization: Bearer ${login.token}
    expect:
      status: 200
```

**Chain reference syntax:**
- `${step_name.field}`: Access top-level field from step response
- `${step_name.nested.field}`: Access nested fields
- Chains execute sequentially and stop on first failure (fail-fast)

## Assertions and Testing

### Status Expectations

```yaml
expect:
  status: 200              # Single status
  status: [200, 201, 204]  # Multiple valid statuses
```

### Body Assertions (JQ Expressions)

```yaml
expect:
  status: 200
  assert:
    # Simple checks
    - .id != null
    - .email != null
    - .active == true

    # Array operations
    - . | length > 0
    - .[0].name != null
    - .[] | .status == "active"  # All items match

    # Type checks
    - . | type == "array"
    - .data | type == "object"

    # Environment variable references
    - .owner.login == env.GITHUB_USER
```

### Header Assertions

```yaml
expect:
  status: 200
  assert:
    headers:
      - .["content-type"] | startswith("application/json")
      - .["x-custom-header"] == "expected-value"
```

## JQ Filtering

Filter and transform response data inline:

```yaml
yapi: v1
url: https://api.example.com/users
method: GET

# Only show specific fields, sorted
jq_filter: '[.[] | {name, email}] | sort_by(.name)'
```

## Protocol Support

### HTTP/REST

```yaml
yapi: v1
url: https://api.example.com/posts
method: POST
content_type: application/json
headers:
  Authorization: Bearer ${TOKEN}
body:
  title: "Hello"
  tags: ["api", "test"]
```

### GraphQL

```yaml
yapi: v1
url: https://countries.trevorblades.com/graphql

graphql: |
  query getCountry($code: ID!) {
    country(code: $code) {
      name
      capital
    }
  }

variables:
  code: "US"
```

### gRPC (with reflection)

```yaml
yapi: v1
url: grpc://localhost:50051
service: helloworld.Greeter
rpc: SayHello

body:
  name: "World"
```

## Request Timeouts

Configure timeouts for HTTP and GraphQL requests using duration strings:

```yaml
yapi: v1
url: https://api.example.com/slow-endpoint
method: GET
timeout: 5s  # Timeout after 5 seconds
```

**Supported duration formats:**
- `"100ms"` - Milliseconds
- `"4s"` - Seconds
- `"1m"` - Minutes
- `"1m30s"` - Combination

**Timeouts in chains:**

Each step in a chain can have its own timeout, and steps inherit the global timeout if not specified:

```yaml
yapi: v1

# Global timeout applies to all steps by default
timeout: 10s

chain:
  # Step 1: Override with shorter timeout
  - name: fast_check
    url: https://api.example.com/health
    timeout: 2s
    expect:
      status: 200

  # Step 2: Uses global timeout (10s)
  - name: normal_request
    url: https://api.example.com/data
    expect:
      status: 200

  # Step 3: Override with longer timeout
  - name: slow_operation
    url: https://api.example.com/process
    timeout: 30s
    expect:
      status: 200
```

**When a timeout occurs:**
- HTTP/GraphQL requests will fail with `context deadline exceeded` error
- The chain will stop execution (fail-fast behavior)
- Use timeouts to prevent hanging on slow or unresponsive endpoints

## Project Structure Best Practices

### Recommended Directory Layout

```
project/
├── yapi.config.yml          # Environment configuration
├── .yapi/                   # All request files in one directory
│   ├── auth/
│   │   ├── login.yapi.yml
│   │   └── refresh.yapi.yml
│   ├── users/
│   │   ├── list-users.yapi.yml
│   │   ├── get-user.yapi.yml
│   │   └── create-user.yapi.yml
│   └── posts/
│       ├── list-posts.yapi.yml
│       └── create-post.yapi.yml
└── tests/
    └── integration/
        ├── auth-flow.yapi.yml
        └── user-workflow.yapi.yml
```

### Example: Web Application (from yapi repo)

```
web/
├── yapi.config.yml          # Manages local/prod environments
└── yapi/                    # Request files for each route
    ├── homepage.yapi.yaml
    ├── blog-what-is-yapi.yapi.yaml
    ├── playground.yapi.yaml
    ├── icon.yapi.yaml
    ├── manifest.yapi.yaml
    └── sitemap.yapi.yaml
```

Each file tests a specific endpoint:

```yaml
# web/yapi/homepage.yapi.yaml
yapi: v1
path: /              # Uses url from yapi.config.yml
method: GET
expect:
  status: 200
```

```yaml
# web/yapi.config.yml
yapi: v1

default_environment: local

environments:
  local:
    url: http://localhost:3000
    vars:
      some_param: default_value

  prod:
    url: https://yapi.run
    vars:
      some_param: some_value
```

### Using Relative Paths

When `yapi.config.yml` provides a base `url`, request files can use relative `path`:

```yaml
# Instead of full URL
url: ${url}/api/v1/users

# Use path (cleaner)
path: /api/v1/users
```

## Commands for LLM Agents to Use

### Running Requests

```bash
# Run a single request
yapi run path/to/request.yapi.yml

# Run with specific environment
yapi run request.yapi.yml -e prod

# Run all requests in directory (test mode)
yapi test ./tests -a

# Watch mode (re-run on file changes)
yapi watch request.yapi.yml

# Stress testing
yapi stress workflow.yapi.yml -n 1000 -p 50
```

### Interactive Mode

```bash
# Launch TUI for fuzzy file selection
yapi
```

## Best Practices for Writing yapi Files

### When Creating HTTP Request Files

1. **Always start with `yapi: v1`**
2. **Use environment variables for sensitive data**: Don't hardcode API keys
3. **Add assertions**: Help users validate responses
4. **Use chains for workflows**: Login → Get Token → Authenticated Request
5. **Leverage JQ filters**: Make output more readable
6. **Prefer relative paths**: When `yapi.config.yml` exists with base URLs

### Example: Creating a Complete API Test

```yaml
yapi: v1
chain:
  # Step 1: Health check
  - name: health
    path: /health
    method: GET
    expect:
      status: 200
      assert:
        - .status == "ok"

  # Step 2: Authenticate
  - name: auth
    path: /api/auth/login
    method: POST
    body:
      email: ${TEST_USER_EMAIL}
      password: ${TEST_USER_PASSWORD}
    expect:
      status: 200
      assert:
        - .token != null
        - .user.email == env.TEST_USER_EMAIL

  # Step 3: Fetch user data
  - name: get_profile
    path: /api/users/me
    method: GET
    headers:
      Authorization: Bearer ${auth.token}
    expect:
      status: 200
      assert:
        - .id != null
        - .email != null
```

## Common Patterns

### Pattern 1: Multi-Environment Setup

```yaml
# yapi.config.yml
yapi: v1
default_environment: dev

defaults:
  vars:
    API_VERSION: v1

environments:
  dev:
    url: http://localhost:8080
    vars:
      DEBUG: "true"

  staging:
    url: https://staging.api.example.com
    vars:
      DEBUG: "false"

  prod:
    url: https://api.example.com
    vars:
      DEBUG: "false"
```

### Pattern 2: Reusable Authentication Chain

```yaml
# auth-flow.yapi.yml
yapi: v1
chain:
  - name: login
    path: /auth/login
    method: POST
    body:
      username: ${USERNAME}
      password: ${PASSWORD}
    expect:
      status: 200
      assert:
        - .access_token != null

  - name: refresh
    path: /auth/refresh
    method: POST
    headers:
      Authorization: Bearer ${login.access_token}
    expect:
      status: 200
```

### Pattern 3: Data-Driven Tests

```yaml
yapi: v1
url: https://jsonplaceholder.typicode.com/users/${USER_ID:-1}
method: GET
expect:
  status: 200
  assert:
    - .id != null
    - .email != null
    - .address.city != null
```

## Tips and Guidelines

1. **Read existing patterns**: Before creating new request files, check existing ones in the project
2. **Follow naming conventions**: Use descriptive names like `get-user.yapi.yml`, not `request1.yapi.yml`
3. **Group related requests**: Use subdirectories (auth/, users/, posts/)
4. **Always add expectations**: Tests without assertions aren't very useful
5. **Use chains for workflows**: Don't create separate files when steps depend on each other
6. **Reference the environment**: Use `${url}` instead of hardcoding base URLs
7. **Add comments**: YAML supports comments - use them to explain complex logic
8. **Keep it simple**: Don't over-engineer - yapi is designed for simplicity

## Error Handling

yapi fails fast on:
- Invalid YAML syntax
- Missing required fields (url, method)
- Failed assertions
- Network errors
- Invalid JQ expressions

When creating files, validate:
- `yapi: v1` is present
- `url` or `path` is defined
- `method` is valid (GET, POST, PUT, PATCH, DELETE)
- JQ assertions are syntactically correct
- Chain references point to valid step names

## CI/CD Integration

yapi integrates seamlessly with GitHub Actions:

```yaml
- uses: jamierpond/yapi/action@0.X.X
  with:
    start: npm run dev
    wait-on: http://localhost:3000/health
    command: yapi test ./tests -a
```

When writing tests for CI/CD:
- Use the `expect` block with assertions
- Group related tests in directories
- Use environment variables for configuration
- Test complete workflows, not just individual endpoints

## Summary

yapi is designed to make API testing version-controllable and automatable. Key principles:

- **Create structured, readable YAML files**
- **Leverage environment configs for multi-environment support**
- **Use chains to test realistic workflows**
- **Add comprehensive assertions**
- **Follow the project's existing patterns**
- **Keep files simple and focused**

The goal is to help users build a git-committable API test suite that serves as both documentation and validation.
