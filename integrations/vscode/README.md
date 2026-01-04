# yapi for VSCode

The API client that lives in your editor. **yapi** is a CLI-first, offline-first, git-friendly API client for HTTP, gRPC, GraphQL, and TCP. This extension provides first-class support for `.yapi.yml` files in VSCode.

## Features

- **Multi-Protocol Support**: HTTP, gRPC, GraphQL, and TCP requests
- **Real-time Validation**: Inline diagnostics with error highlighting
- **Live Response Panel**: Split-pane view with syntax-highlighted JSON responses
- **Quick Examples**: Insert example configurations with one command
- **Keyboard Shortcuts**: `Cmd+Enter` / `Ctrl+Enter` to run requests
- **Execution Timing**: See request completion times

## Requirements

- [yapi CLI](https://github.com/jamierpond/yapi) must be installed and available in your PATH

Install yapi:
```bash
# macOS
curl -fsSL https://yapi.run/install/mac.sh | bash

# Linux
curl -fsSL https://yapi.run/install/linux.sh | bash

# Or with Homebrew
brew tap jamierpond/yapi && brew install --cask yapi
```

## Usage

1. Create a `.yapi.yml` or `.yapi.yaml` file
2. Write your API request configuration
3. Press `Cmd+Enter` / `Ctrl+Enter` or click the "Run" button in the toolbar
4. View the response in the side panel

### Example

```yaml
# hello.yapi.yml
yapi: v1
url: https://httpbin.org/post
method: POST
content_type: application/json

body:
  message: "Hello from yapi"
  timestamp: "2024-01-01"
```

> **Note:** The `yapi: v1` version tag is required at the top of all config files.

## Commands

- **yapi: Run yapi** - Execute the current yapi file (`Cmd+Enter` / `Ctrl+Enter`)
- **yapi: Insert Example** - Quick insert example configurations
- **yapi: Restart Language Server** - Restart the yapi LSP server

## Extension Settings

- `yapi.executablePath`: Path to the yapi executable (default: `yapi` - searches in PATH)

## Validation

The extension provides real-time validation for yapi files:

- Missing required fields (e.g., `url`)
- Conflicting fields (e.g., both `body` and `json`)
- Protocol-specific requirements (e.g., gRPC needs `service` and `rpc`)
- YAML syntax errors with line-level diagnostics

## Keyboard Shortcuts

- `Cmd+Enter` / `Ctrl+Enter` - Run the current yapi file

## Supported Protocols

### HTTP/REST
```yaml
yapi: v1
url: https://api.example.com
method: POST
path: /users
content_type: application/json
body:
  name: "John Doe"
```

### gRPC
```yaml
yapi: v1
url: grpc://localhost:50051
service: helloworld.Greeter
rpc: SayHello
body:
  name: "yapi User"
```

### TCP
```yaml
yapi: v1
url: tcp://localhost:9877
method: tcp
data: "Hello!\n"
encoding: text
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
  code: "BR"
```

## Release Notes

### 0.0.1

Initial release with:
- Multi-protocol support (HTTP, gRPC, GraphQL, TCP)
- Real-time validation
- Live response panel
- Example snippets
- Keyboard shortcuts
