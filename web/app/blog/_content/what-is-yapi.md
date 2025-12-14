# What is yapi?

## Yapi is the API client that runs in your terminal.

Yapi is the hackers Postman, Insomnia, Bruno.

It is a command line tool that makes it easy to interact with APIs from your terminal.


This request:
```yaml
yapi: v1
url: https://api.github.com/repos/jamierpond/yapi
method: GET
jq_filter: '. | {stars: .stargazers_count, name: .name}'
```

Gives you this response:
```json
{
  "name": "yapi"
  "stars": 420,
}
```

## Why make another API client?
Well it started just from me chaining a few CLI tools together with bash.

```bash

```
