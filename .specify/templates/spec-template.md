# Feature Specification: [FEATURE NAME]

**Branch**: `[###-feature-name]` | **Created**: [DATE] | **Status**: Draft

## Overview

[1-2 sentences describing what this feature does]
<!--
    Note 'users' can be end users of yapi, or the developers who maintain it.
    Make sure this is clear.
-->

## User Stories

### US1 - [Title] (P1)

[User journey in plain language]

**CLI Usage**:
```bash
yapi [command] [args]
```

**Acceptance**:
- Given [state], when [action], then [outcome]

---

### US2 - [Title] (P2)

[User journey]

**CLI Usage**:
```bash
yapi [command] [args]
```

**Acceptance**:
- Given [state], when [action], then [outcome]

## Requirements

### Functional

- **FR-001**: `yapi` MUST [capability]
- **FR-002**: `yapi` MUST [capability]

### YAML Schema (if applicable)

```yaml
yapi: v1
# New fields this feature adds
[field]: [type/example]
```

### Protocol Support

| Protocol | Supported | Notes |
|----------|-----------|-------|
| HTTP     | [ ]       |       |
| gRPC     | [ ]       |       |
| GraphQL  | [ ]       |       |
| TCP      | [ ]       |       |

## Edge Cases

- What happens when [boundary condition]?
- How does `yapi` handle [error scenario]?

## Success Criteria

- [ ] Feature works via CLI without GUI
- [ ] Configuration stored in `.yapi.yml` files
- [ ] Works across all applicable protocols
- [ ] Minimal implementation, no unnecessary complexity
- [ ] Tests pass: `make test`
- [ ] Lint passes: `make lint`
