# Protocols

yapi supports HTTP, gRPC, GraphQL, and TCP as first-class protocols.
Transport is detected from the URL scheme and config fields.

## HTTP / REST

The default protocol. Standard request/response:

```yaml
yapi: v1
url: https://api.example.com/users
method: POST
content_type: application/json
headers:
  Authorization: Bearer ${TOKEN}
body:
  name: "Alice"
  email: "alice@example.com"
expect:
  status: 201
```

### Form Data

```yaml
yapi: v1
url: https://example.com/login
method: POST
form:
  username: admin
  password: ${PASSWORD}
```

## GraphQL

Detected by the `graphql:` field:

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

GraphQL requests are always POST. The `graphql` field contains your query or mutation,
and `variables` holds the GraphQL variables.

## gRPC

Detected by `grpc://` URL scheme. Requires server reflection or proto files:

```yaml
yapi: v1
url: grpc://localhost:50051
service: helloworld.Greeter
rpc: SayHello

body:
  name: "World"
```

### With Proto Files

```yaml
yapi: v1
url: grpc://localhost:50051
service: helloworld.Greeter
rpc: SayHello
proto: ./proto/helloworld.proto
proto_path: ./proto

body:
  name: "World"
```

### Insecure / Plaintext

```yaml
plaintext: true    # No TLS
insecure: true     # Skip TLS verification
```

## TCP

Detected by `tcp://` URL scheme. Raw socket communication:

```yaml
yapi: v1
url: tcp://localhost:9877
data: '{"type":"health","params":{}}'
encoding: text           # text (default), hex, base64
read_timeout: 5          # Seconds to wait for response
idle_timeout: 500        # Milliseconds before considering response complete
close_after_send: false  # Keep connection open to read response
```

## Auto-Detection Summary

| Indicator | Protocol |
|---|---|
| `grpc://` URL | gRPC |
| `tcp://` URL | TCP |
| `graphql:` field present | GraphQL |
| Everything else | HTTP |

## See Also

- `yapi docs send` — Quick one-off requests (auto-detects protocol)
- `yapi docs config` — Full field reference for all protocols
