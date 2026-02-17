# JQ Filtering

JQ lets you filter and transform JSON responses inline —
both for display (`jq_filter`) and for validation (`assert`).

## Response Filtering

Use `jq_filter` to transform what gets displayed:

```yaml
yapi: v1
url: https://api.example.com/users
method: GET
jq_filter: '[.[] | {name, email}] | sort_by(.name)'
```

This filters the response before printing, so you only see what matters.

## JQ in Assertions

Assertions are JQ expressions that must evaluate to `true`:

```yaml
expect:
  assert:
    - .id != null
    - . | length > 0
    - .data | type == "object"
    - .[0].name | startswith("A")
```

## Common Patterns

### Check array length
```yaml
assert:
  - . | length > 0
  - .items | length == 10
```

### Check type
```yaml
assert:
  - . | type == "array"
  - .data | type == "object"
  - .count | type == "number"
```

### String operations
```yaml
assert:
  - .name | startswith("test_")
  - .email | endswith("@example.com")
  - .url | contains("api")
```

### Check all items in array
```yaml
assert:
  - .[] | .status == "active"     # Every item matches
  - [.[] | .score > 0] | all      # All scores positive
```

### Select specific fields
```yaml
jq_filter: '{id, name, email}'         # Single object
jq_filter: '[.[] | {id, name}]'        # Array of objects
```

### Sort and limit
```yaml
jq_filter: 'sort_by(.created_at) | reverse | .[:5]'
```

## JQ with yapi send

Apply JQ filters on the command line:

```bash
yapi send https://api.example.com/users --jq '.[0].name'
```

## See Also

- `yapi docs assert` — Assertions on status, body, and headers
- `yapi docs send` — Quick one-off requests with --jq flag
