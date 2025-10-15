# yapi - YAML API Testing

> Because writing curl commands by hand is so 2015

## What is this?

**yapi** (pronounced "yappy", like an excited dog) is a dead-simple API testing tool that uses YAML config files instead of making you memorize curl flags or write bloated test frameworks.

Write your API calls in YAML. Run them with one command. Get pretty JSON output. That's it.

```bash
# Interactive mode with fzf
./yapi.sh

# Or be explicit
./yapi.sh -c examples/create-post.yapi.yml

# Override the URL (great for testing local vs production)
./yapi.sh -c my-test.yapi.yml -u http://localhost:3000
```

## Why?

- **No more curl archaeology** - Stop scrolling through bash history trying to find that one curl command from 3 weeks ago
- **Git-trackable tests** - Your API tests live in version control, right next to your code
- **Dead simple** - If you can write YAML, you can test APIs
- **fzf integration** - Interactive file picker because typing paths is for chumps
- **JSON pretty-printing** - Automatic `jq` formatting because we're not savages

## Installation

### Requirements

- `bash` (you probably have this)
- `curl` (you definitely have this)
- `yq` - [Install it](https://github.com/mikefarah/yq)
- `fzf` - [Install it](https://github.com/junegunn/fzf) (optional, but highly recommended)
- `jq` - For pretty JSON output (optional but nice)

### Setup

```bash
# Clone it
git clone <your-repo-url>
cd yapi

# Make it executable
chmod +x yapi.sh

# Optional: symlink it somewhere in your PATH
ln -s $(pwd)/yapi.sh /usr/local/bin/yapi

# Now you can run it from anywhere
yapi -c my-test.yapi.yml
```

## Creating Your First Test

Create a file called `hello-world.yapi.yml`:

```yaml
# yaml-language-server: $schema=https://pond.audio/yapi/schema

url: https://httpbin.org
path: /get
method: GET
```

Run it:

```bash
./yapi.sh -c hello-world.yapi.yml
```

Boom. You just tested an API.

## More Examples

### Simple GET request

```yaml
# yaml-language-server: $schema=https://pond.audio/yapi/schema

url: https://api.github.com
path: /users/octocat
method: GET
```

### POST with YAML body

When you want to write your JSON payload in YAML (because YAML is nicer):

```yaml
# yaml-language-server: $schema=https://pond.audio/yapi/schema

url: https://httpbin.org
path: /post
method: POST
content_type: application/json

body:
  title: "My awesome post"
  description: "Testing yapi like a boss"
  userId: 123
  tags:
    - "testing"
    - "api"
  metadata:
    source: "yapi"
    version: "1.0"
```

The `body` field gets automatically converted from YAML to JSON. Magic!

### POST with raw JSON literal

Sometimes you just want to paste JSON as-is (like when copying from docs):

```yaml
# yaml-language-server: $schema=https://pond.audio/yapi/schema

url: https://httpbin.org
path: /post
method: POST
content_type: application/json

json: |
  {
    "id": 123,
    "status": "active",
    "nested": {
      "foo": "bar"
    }
  }
```

Use `json` for raw JSON, use `body` for YAML that gets converted. Pick your poison.

## Pro Tips

### Use fzf for maximum laziness

Just run `./yapi.sh` with no arguments and fzf will let you pick from all your `*.yapi.yml` files:

```bash
./yapi.sh
# Shows you a fuzzy-searchable list of all your test files
```

### File naming convention

By default, yapi searches for files matching `*.yapi.yml` or `*.yapi.yaml` in your git repo. This keeps them separate from other YAML files.

Want to search ALL yaml files? Use the `-a` flag:

```bash
./yapi.sh -a
```

### Override URLs for different environments

```bash
# Test against production
./yapi.sh -c user-login.yapi.yml -u https://api.production.com

# Test against local dev
./yapi.sh -c user-login.yapi.yml -u http://localhost:3000

# Test against staging
./yapi.sh -c user-login.yapi.yml -u https://api.staging.com
```

Same test file, different environments. Beautiful.

### Schema validation

Add this line to the top of your `.yapi.yml` files for autocomplete and validation in VS Code:

```yaml
# yaml-language-server: $schema=https://pond.audio/yapi/schema
```

Now your editor will yell at you if you mess up the format.

## YAML Schema Reference

```yaml
# yaml-language-server: $schema=https://pond.audio/yapi/schema

# Base URL (can be overridden with -u flag)
url: https://api.example.com

# Path to append to URL
path: /api/v1/users

# HTTP method (default: GET)
method: POST

# Content type (required when using body or json)
content_type: application/json

# Option 1: YAML body (gets converted to JSON)
body:
  key: value
  nested:
    stuff: here

# Option 2: Raw JSON literal (use either body OR json, not both)
json: |
  {
    "key": "value"
  }
```

## Examples Directory

Check out the [`examples/`](examples/) directory for more:

- [`google.yapi.yml`](examples/google.yapi.yml) - Simple GET request
- [`create-post.yapi.yml`](examples/create-post.yapi.yml) - POST with YAML body
- [`json-literal.yapi.yml`](examples/json-literal.yapi.yml) - POST with raw JSON

## Command Line Options

```
Usage: yapi.sh [OPTIONS]

Options:
  -c, --config FILE    Path to YAML config file (required, or uses fzf)
  -u, --url URL        Override base URL from config file
  -a, --all            Search all YAML files (default: git-tracked *.yapi.yml only)
  -h, --help           Display help message

Examples:
  yapi.sh -c test.yaml
  yapi.sh --config test.yaml --url http://localhost:8080
  yapi.sh --all
```

## FAQ

**Q: Why not just use Postman/Insomnia/etc?**
A: Those are great, but they're GUI apps with proprietary formats. yapi files are just YAML in git. Diff them, PR them, CI/CD them.

**Q: Why not use [insert fancy testing framework]?**
A: Sometimes you just want to hit an endpoint and see what it returns. No setup, no config, no ceremony.

**Q: Can I use this in CI/CD?**
A: Sure! It's just a bash script. Add some response validation and you're golden.

**Q: What about auth headers, cookies, etc?**
A: Coming soon! For now, you can fork and extend the script.

**Q: The name is weird.**
A: That's not a question. But yes, it is.

## Contributing

Found a bug? Want a feature? PRs welcome!

## License

Do whatever you want with it. YOLO.

---

Made with questionable decisions and too much coffee.
