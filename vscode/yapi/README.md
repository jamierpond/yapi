# yapi for VSCode

YAML API Testing Tool for VSCode - supports HTTP/REST, gRPC, and TCP protocols.

## Features

- **Multi-Protocol Support**: HTTP, gRPC, and TCP requests
- **Real-time Validation**: Inline diagnostics with error highlighting
- **Live Response Panel**: Split-pane view with syntax-highlighted JSON responses
- **Quick Examples**: Insert example configurations with one command
- **Hot Reload**: Auto-run on save (configurable)
- **Keyboard Shortcuts**: `Cmd+Enter` / `Ctrl+Enter` to run requests
- **Execution Timing**: See request completion times

## Requirements

- [yapi CLI](https://github.com/jamierpond/yapi) must be installed and available in your PATH
- For gRPC: `grpcurl` (optional, for gRPC requests)
- For TCP: `nc` / `netcat` (usually pre-installed)

## Usage

1. Create a `.yapi.yml` or `.yapi.yaml` file
2. Write your API request configuration
3. Press `Cmd+Enter` / `Ctrl+Enter` or click the "Run" button in the toolbar
4. View the response in the side panel

### Example

```yaml
# hello.yapi.yml
url: https://httpbin.org/post
method: POST
content_type: application/json

body:
  message: "Hello from yapi"
  timestamp: "2024-01-01"
```

## Commands

- **yapi: Run yapi** - Execute the current yapi file (`Cmd+Enter` / `Ctrl+Enter`)
- **yapi: Insert Example** - Quick insert example configurations

## Extension Settings

- `yapi.autoRunOnSave`: Automatically run yapi when a .yapi.yml file is saved (default: `true`)
- `yapi.runOnSave`: Run yapi on `Cmd+S` / `Ctrl+S` instead of just saving (default: `false`)

## Validation

The extension provides real-time validation for yapi files:

- Missing required fields (e.g., `url`)
- Conflicting fields (e.g., both `body` and `json`)
- Protocol-specific requirements (e.g., gRPC needs `service` and `rpc`)
- YAML syntax errors with line-level diagnostics

## Keyboard Shortcuts

- `Cmd+Enter` / `Ctrl+Enter` - Run the current yapi file
- `Cmd+S` / `Ctrl+S` - Save (or run if `yapi.runOnSave` is enabled)

## Supported Protocols

### HTTP/REST
```yaml
url: https://api.example.com
method: POST
path: /users
content_type: application/json
body:
  name: "John Doe"
```

### gRPC
```yaml
url: grpc://localhost:50051
service: hello.HelloService
rpc: SayHello
plaintext: true
body:
  greeting: "World"
```

### TCP
```yaml
url: tcp://localhost:9877
method: tcp
data: "Hello!\n"
encoding: text
```

## Release Notes

### 0.0.1

Initial release with:
- Multi-protocol support (HTTP, gRPC, TCP)
- Real-time validation
- Live response panel
- Example snippets
- Keyboard shortcuts
