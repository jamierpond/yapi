<!--
  Sync Impact Report
  ===================
  Version change: 1.2.0 -> 1.3.0

  Modified principles: N/A

  Added sections:
  - Core Principles V. Dogfooding (1.1.0)
  - Core Principles VI. Minimal Code (1.2.0)
  - Core Principles VII. Single Code Path (1.3.0)

  Removed sections: N/A

  Templates requiring updates:
  - .specify/templates/plan-template.md: Add Dogfooding to Constitution Check table [DONE]
  - .specify/templates/plan-template.md: Add Minimal Code to Constitution Check table
  - .specify/templates/plan-template.md: Add Single Code Path to Constitution Check table

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

**Rationale**: Less code means fewer bugs, faster builds, easier onboarding, and reduced
maintenance burden. The best code is no code at all.

### VII. Single Code Path

The LSP and CLI MUST use identical code paths for config parsing, validation, and
compilation. The `compiler` package is the single source of truth.

- LSP MUST NOT duplicate logic that exists in core packages (config, validation, compiler)
- Changes to execution logic MUST NOT require parallel changes in the LSP
- The langserver is a thin adapter layer: it converts LSP protocol to yapi core calls
- If you're writing validation/parsing logic in langserver, you're doing it wrong
- All analyzer functions MUST be usable by both CLI and LSP with the same parameters
- New features affecting config interpretation MUST be implemented in core, not langserver

**Rationale**: Divergent code paths cause bugs. The LSP showing different behavior than
CLI execution is a critical defect. We just fixed a bug where env_files worked in CLI
but not LSP because they used different analyzer calls. Never again.

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

**Version**: 1.3.0 | **Ratified**: 2025-10-14 | **Last Amended**: 2026-01-03
