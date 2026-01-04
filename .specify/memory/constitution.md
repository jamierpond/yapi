<!--
  Sync Impact Report
  ===================
  Version change: 1.3.1 -> 1.4.0

  Modified principles: N/A

  Added sections:
  - Core Principles V. Dogfooding (1.1.0)
  - Core Principles VI. Minimal Code (1.2.0)
  - Core Principles VII. Single Code Path (1.3.0)
  - Core Principles VIII. Fail Fast (1.4.0)

  Removed sections: N/A

  Templates requiring updates:
  - .specify/templates/plan-template.md: Add Dogfooding to Constitution Check table [DONE]
  - .specify/templates/plan-template.md: Add Minimal Code to Constitution Check table
  - .specify/templates/plan-template.md: Add Single Code Path to Constitution Check table
  - .specify/templates/plan-template.md: Add Fail Fast to Constitution Check table

  Follow-up TODOs: None
-->

# yapi Constitution

## Core Principles

### I. CLI-First Design

Every feature in yapi MUST be accessible via the command line. The terminal is the
primary interface.

- All functionality MUST work without a GUI
- Text input/output protocol: stdin/args -> stdout, errors -> stderr
- Interactive TUI features are optional enhancements, not requirements
- Support both JSON and human-readable output formats
- Every command MUST be scriptable and pipeable

**Rationale**: yapi exists because heavy Electron apps are not developer-friendly.
If a feature cannot be used from a terminal, it violates the core mission.

### II. Git-Friendly

All yapi configuration and request definitions MUST be stored in plain-text YAML
files that can be committed, reviewed, and diffed.

- Request files use `.yapi.yml` extension for discoverability
- No binary formats or database storage for user configurations
- Configuration MUST NOT require external services to function
- Environment-specific values MUST be injectable via environment variables or files
- Version control workflows (branching, merging, reviewing) MUST work naturally

**Rationale**: API definitions are code. They belong in version control alongside
the services they test, not locked in proprietary formats or cloud services.

### III. Protocol Agnostic

yapi MUST treat HTTP, gRPC, GraphQL, and TCP as first-class citizens with a unified
configuration model.

- All protocols share the same YAML structure where applicable
- Protocol-specific features are additive, not replacements
- New protocol support MUST NOT break existing request files
- Request chaining and assertions work across all protocols
- Environment configuration applies uniformly to all protocols

**Rationale**: Modern APIs use multiple protocols. Developers should not need
different tools for different protocols.

### IV. Simplicity

Start with the simplest implementation. Complexity MUST be justified.

- YAGNI: Do not add features "just in case"
- Single-purpose files and functions
- No unnecessary abstractions or indirection layers
- Configuration options MUST have sensible defaults
- A minimal `.yapi.yml` file MUST work without boilerplate
- Prefer explicit over implicit behavior

**Rationale**: Complexity accumulates. Every abstraction adds cognitive load.
The default path should be the simple path.

### V. Dogfooding

The yapi webapp MUST use yapi itself for all API interactions where feasible.

- New webapp features SHOULD be built using yapi request files
- Internal API testing MUST use yapi, not external tools
- The webapp serves as a live demonstration of yapi capabilities
- Friction discovered while dogfooding MUST inform CLI/core improvements
- If a workflow is painful in the webapp, it's painful for users too
- **`apps/web/yapi/` files MUST demonstrate best practices**: These files are living examples. When syntax or features change, update these files in the same PR. They should always showcase the latest yapi capabilities.

**Rationale**: Eating our own dog food exposes usability issues before users hit them.
The webapp is both a product and a continuous integration test for yapi itself.

### VI. Minimal Code

You can't have bugs in code you don't have. Every line of code is a liability.

- Actively seek opportunities to delete code rather than add it
- New features SHOULD reduce total LOC when possible through consolidation
- Prefer removing unused code over commenting it out
- Duplication is acceptable if the alternative is a complex abstraction
- Before adding a dependency, consider if the functionality can be achieved with less code
- Refactoring that increases LOC requires explicit justification
- **NO DEPRECATED FUNCTIONS**: When you change an API, update all callers and DELETE the old function in the same PR. Never mark internal functions as "Deprecated" - just delete them.
- **NO BACKWARDS-COMPAT SHIMS**: No `_unused` renames, no re-exports, no wrapper functions "for compatibility". If it's unused, it's deleted.
- **DELETE TESTS FOR DELETED CODE**: When you delete a function, delete its tests. Dead tests are dead weight.

**Rationale**: This project is maintained by one person. Every line of code is a maintenance burden.
Less code means fewer bugs, faster builds, easier onboarding. The best code is no code at all.

### VII. One Pipeline, Different Stopping Points

ALL yapi commands are the SAME pipeline. They differ only in where they stop and how
many times they iterate. This is not a guideline - it is the architecture.

```
parse → validate → execute
  ↑         ↑          ↑
  │         │          └── run stops here (once)
  │         │              test stops here (N files)
  │         │              stress stops here (N×M times)
  │         │              watch stops here (on every file change)
  │         │
  │         └── validate stops here (once)
  │             lsp stops here (on every keystroke)
  │
  └── All commands start here
```

**The Rules:**

- `validate` = parse + validate. That's it. One file, one pass, return diagnostics.
- `run` = validate + execute. One file, one pass, return result.
- `test` = run, but for N files. Same pipeline, iterated.
- `stress` = run, but N times concurrently. Same pipeline, parallelized.
- `watch` = run, but re-triggered on file changes. Same pipeline, looped.
- `lsp` = validate, but persistent and re-triggered on edits. Same pipeline, streaming.

**Implementation Requirements:**

- ALL commands MUST call `validation.Analyze()` for parsing and validation
- ALL commands that execute MUST go through `core.Engine.RunConfig()`
- The langserver is a thin adapter: LSP protocol → `validation.Analyze()` → LSP response
- There is NO command-specific parsing logic. NONE. EVER.
- If you're writing validation code anywhere except `internal/validation`, you're wrong
- If `validate` and `run` can produce different diagnostics for the same file, that's a bug
- If the LSP shows different errors than `yapi validate`, that's a critical defect

**What This Means In Practice:**

- Adding a new validation rule? Add it to `validation.Analyze()`. Done. All commands get it.
- Fixing a parse bug? Fix it in `config/`. Done. All commands get it.
- Adding environment variable support? Add it to the core resolver. Done. All commands get it.
- NEVER add command-specific preprocessing, postprocessing, or "special cases"

**Rationale**: We don't want command-specific bugs. We don't want to maintain multiple code paths. Keep it simple. Keep it DRY. No one wants wet code.

### VIII. Fail Fast

Code MUST fail immediately and loudly when something is wrong. Silent failures and deferred errors are bugs.

- Use assertions liberally. If a condition should never happen, assert it.
- Validate inputs at the boundary. Reject garbage immediately, don't propagate it.
- Panic on impossible states rather than returning meaningless defaults.
- Error messages MUST be specific: what failed, why, and where.
- NO defensive coding that papers over bugs. If caller passes nil, panic. Don't check and silently return.
- NO "graceful degradation" that hides broken behavior. If it's broken, STOP.
- Prefer hard crashes over corrupted state. A crash is debuggable. Corrupted data is a nightmare.
- Tests MUST assert behavior, not just "run without error". A test that doesn't assert is not a test.

**Rationale**: The earlier you find a bug, the cheaper it is to fix. Assertions and hard failures surface bugs at development time, not in production. Whimsy code that "handles" errors by ignoring them creates debugging nightmares. Fail hard, fail fast, fix it now.

## Quality Standards

### Testing Requirements

- New features MUST include tests
- Bug fixes SHOULD include regression tests
- Integration tests MUST cover multi-protocol scenarios
- Performance-critical paths MUST have benchmarks

### Code Quality

- All Go code MUST pass `make lint`
- All code MUST pass `make test`
- Public APIs MUST have documentation
- Error messages MUST be actionable (tell users how to fix the problem)

### Documentation

- README MUST be kept up to date with new features
- Examples directory MUST contain working samples for all protocols
- Breaking changes MUST be documented in release notes

## Development Workflow

### Feature Development

1. Create feature branch from main
2. Write failing tests (where applicable)
3. Implement feature
4. Ensure `make build && make test && make lint` passes
5. Update documentation and examples
6. Submit for review

### Release Process

- Semantic versioning: MAJOR.MINOR.PATCH
- MAJOR: Breaking changes to CLI or YAML schema
- MINOR: New features, backward compatible
- PATCH: Bug fixes, documentation updates

### Review Standards

- All changes MUST pass CI checks
- Breaking changes require explicit acknowledgment
- Performance regressions require justification

## Governance

This constitution supersedes all other development practices for the yapi project.
All contributions MUST comply with these principles.

### Amendment Process

1. Propose amendment with rationale
2. Document impact on existing features
3. Update affected templates and documentation
4. Increment constitution version appropriately

### Versioning Policy

- MAJOR: Backward-incompatible principle changes or removals
- MINOR: New principles or materially expanded guidance
- PATCH: Clarifications, wording improvements

### Compliance

- All pull requests MUST be reviewed against these principles
- Complexity MUST be justified in PR descriptions
- Principle violations require explicit exemption with documented rationale

**Version**: 1.4.0 | **Ratified**: 2025-10-14 | **Last Amended**: 2026-01-04
