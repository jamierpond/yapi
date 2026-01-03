<!--
  Sync Impact Report
  ===================
  Version change: N/A -> 1.0.0 (initial constitution)

  Modified principles: N/A (initial)

  Added sections:
  - Core Principles (4 principles)
  - Quality Standards
  - Development Workflow
  - Governance

  Removed sections: N/A

  Templates requiring updates:
  - .specify/templates/plan-template.md: N/A (no constitution-specific references)
  - .specify/templates/spec-template.md: N/A (no constitution-specific references)
  - .specify/templates/tasks-template.md: N/A (no constitution-specific references)
  - .specify/templates/agent-file-template.md: N/A (generic template)
  - .specify/templates/checklist-template.md: N/A (generic template)

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

**Version**: 1.0.0 | **Ratified**: 2025-10-14 | **Last Amended**: 2026-01-03
