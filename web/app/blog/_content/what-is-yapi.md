# What is yapi?
## Yapi is the API client that runs in your terminal.

Yapi is the hacker's Postman, Insomnia, Bruno.

It is a command line tool that makes it easy to interact with APIs from your
terminal.

### Yapi speaks HTTP
I wanted a fun way to make HTTP requests from the terminal (without massive `curl` incantations).
#### GET
This request:
```yaml
yapi: v1
method: GET
url: https://api.github.com/search/repositories
headers:
  Authorization: Bearer ${GITHUB_PAT}
query:
  q: yapi in:name, jamierpond in:owner
jq_filter: |
    .items[] | {
      name: .name,
      stars: .stargazers_count,
      url: .html_url
    }
```

Gives you this response:
```json
{
  "name": "yapi",
  "stars": 5, // at time of writing!
  "url": "https://github.com/jamierpond/yapi"
}
{
  "name": "yapi-blog",
  "stars": 0,
  "url": "https://github.com/jamierpond/yapi-blog"
}
{
  "name": "homebrew-yapi",
  "stars": 0,
  "url": "https://github.com/jamierpond/homebrew-yapi"
}
```

#### POST
This request:
```yaml
yapi: v1
url: https://httpbin.org/post
method: POST
body:
  title: "Hello from yapi"
jq_filter: ".json"
```
Gives you this response:
```json
{
  "title": "Hello from yapi"
}
```
### Yapi speaks gRPC

#### Unary RPC
This request:
```yaml
yapi: v1
url: grpc://grpcb.in:9000
service: hello.HelloService
rpc: SayHello
plaintext: true
body:
  greeting: "World"
```
Gives you this response:
```yaml
{
  "reply": "hello World"
}
```


You can also do PUT, PATCH, DELETE and any other HTTP method.

## Why make another API client?
Well it started just from me chaining a few CLI tools together with bash.

This was literally all yapi v0 was, feel free to go back in the git history and see for yourself!

```bash
#!/bin/bash
set -e

config="$1"
url="$2"

default_url="http://localhost:3000"
usage_string="Usage: $0 <config> <url=$default_url>"
if [ -z "$config" ]; then
  echo "$usage_string"
  exit 1
fi

if [ -z "$url" ]; then
  url="$default_url"
fi

config_exists=$(yq e 'true' $config 2>/dev/null || echo "false")
if [ "$config_exists" != "true" ]; then
  echo "Config file $config does not exist or is not a valid YAML file."
  exit 1
fi

endpoint=$(yq e '.endpoint' $config)
json=$(yq e '.json' $config)
method=$(yq e '.method' $config)

url="${url%/}$endpoint"

curl
  -X "$method"
  "$url"
  -H "Content-Type: application/json"
  -d "$json"
  -s | jq
```


Then I wanted to add more ergonomics, like fuzzy finding your yapi files on disk...
```bash

```
