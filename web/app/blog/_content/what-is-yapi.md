# (Draft) What is yapi?
## Yapi is the API client that runs in your terminal.

> Yapi is the hacker's Postman, Insomnia or Bruno.

Yapi is an OSS command line tool that makes it easy to test APIs from your terminal. Yapi speaks HTTP, gRPC, TCP, GraphQL (and more coming soon).

### Yapi speaks HTTP
I wanted a fun way to make HTTP requests from the terminal (without massive, ad-hoc `curl` incantations).
#### GET
This request:
```yaml
# search.yapi.yml
yapi: v1
method: GET
url: https://api.github.com/search/repositories
headers:
  Authorization: Bearer ${GITHUB_PAT} # Reads from your environment
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
yapi run search.yapi.yml
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
# create-issue.yapi.yml
yapi: v1
method: POST
url: https://api.github.com/repos/jamierpond/yapi/issues
headers:
  Accept: application/vnd.github+json
  Authorization: Bearer ${GITHUB_PAT}
body:
  title: Help yapi made me too productive.
  body: |
    Now I can't stop YAPPIN' about yapi!
```
Gives you this response:
```json
yapi run create-issue.yapi.yml
{
  "active_lock_reason": null,
  "assignee": null,
  "assignees": [],
  "author_association": "OWNER",
  "body": "Now I can't stop YAPPIN' about yapi!\n",
  "closed_at": null,
  "closed_by": null,
  "comments": 0,
  // ...blah blah blah
}
```
You can also do PUT, PATCH, DELETE and any other HTTP method.

### Yapi supports chaining requests between protocols
#### Multi-protocol chaining
Yapi makes it easy to chain requests and share data between them, even if they are different protocols.
```yaml
# multi-protocol-chain.yapi.yml
yapi: v1
chain:
  - name: get_todo
    url: https://jsonplaceholder.typicode.com/todos/1
    method: GET

  - name: tcp_echo
    url: tcp://tcpbin.com:4242
    data: "Todo: ${get_todo.title}\n"
    encoding: text
    read_timeout: 5
    close_after_send: true

  - name: grpc_hello
    url: grpc://grpcb.in:9000
    service: hello.HelloService
    rpc: SayHello
    plaintext: true
    body:
      greeting: $get_todo.title

  - name: create_post
    url: https://jsonplaceholder.typicode.com/posts
    method: POST
    headers:
      Content-Type: application/json
    body:
      original_todo: $get_todo.title
      grpc_reply: $grpc_hello.reply
      userId: $get_todo.userId
    expect:
      status: 200
      assert:
        # run tests using jq assertions
        - .userId == $get_todo.userId
```

And gives you this response:
```json
yapi run multi-protocol-chain.yapi.yml
{
  "completed": false,
  "id": 1,
  "title": "delectus aut autem",
  "userId": 1
}
{
  "reply": "hello delectus aut autem"
}
{
  "grpc_reply": "hello delectus aut autem",
  "id": 101,
  "original_todo": "delectus aut autem",
  "userId": 1
}
```


