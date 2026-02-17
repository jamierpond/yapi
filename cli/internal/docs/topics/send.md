# Send

`yapi send` makes quick one-off requests without a config file.
Think of it as curl with better defaults.

## Basic Usage

```bash
yapi send https://httpbin.org/get
yapi send https://httpbin.org/post '{"hello":"world"}'
```

## Method Detection

- **No body**: defaults to GET
- **Body provided**: defaults to POST
- **Override with -X**: `yapi send -X PUT https://api.example.com/users/1 '{"name":"Bob"}'`

## Headers

```bash
yapi send -H "Authorization: Bearer token123" https://api.example.com/me
yapi send -H "Content-Type: text/plain" -H "X-Custom: value" https://example.com/data
```

## JQ Filtering

Filter the response with `--jq`:

```bash
yapi send https://api.example.com/users --jq '.[0].name'
yapi send https://api.example.com/users --jq '[.[] | {name, email}]'
```

## JSON Output

Get structured JSON output with metadata:

```bash
yapi send https://httpbin.org/get --json
```

## Protocol Auto-Detection

The URL scheme determines the protocol:

```bash
yapi send https://api.example.com/users          # HTTP
yapi send tcp://localhost:9877 '{"type":"ping"}'  # TCP
yapi send grpc://localhost:50051                  # gRPC
```

## Verbose Mode

See request and response details:

```bash
yapi send -v https://httpbin.org/get
```

## Examples

```bash
# GET request
yapi send https://jsonplaceholder.typicode.com/posts/1

# POST with JSON body
yapi send https://httpbin.org/post '{"key":"value"}'

# PUT with headers
yapi send -X PUT https://api.example.com/items/1 '{"name":"updated"}' \
  -H "Authorization: Bearer ${TOKEN}"

# Filter response
yapi send https://jsonplaceholder.typicode.com/users --jq '.[].name'

# TCP raw message
yapi send tcp://localhost:9877 '{"type":"health","params":{}}'
```

## See Also

- `yapi docs protocols` — Protocol-specific details
- `yapi docs jq` — JQ filtering and expressions
