# Feature Specification: Next Branch Workflow

**Branch**: `007-jp-next-branch` | **Created**: 2026-01-04 | **Status**: Draft

## Overview

Establish a staging branch named `next` that serves as an unstable/nightly integration point between feature branches and the main production branch. The `next` branch enables developers to validate changes before merging to main, and automatically produces nightly releases on every push.

## Clarifications

### Session 2026-01-04

- Q: When should next pre-releases trigger? → A: Every push to `next`
- Q: What version format for next releases? → A: `v0.X.Y-next.<short-hash>` (semver compliant, 7-char hash)
- Q: What subdomain for next web app? → A: Single `next.yapi.run` (always latest push)
- Q: What artifacts for next pre-releases? → A: CLI binaries + VS Code extension (no Homebrew)
- Q: How is next base version determined? → A: Latest stable tag (e.g., v0.5.2-next.abc1234)
- Q: VS Code extension next releases? → A: Yes - included in same GitHub pre-release as CLI binaries

## User Stories

### US1 - Merge Feature to Next (P1)

A developer completes work on a feature branch and wants to validate it in a staging environment before merging to main. They merge their feature branch into `next`, which triggers CI/CD pipelines and produces a nightly release.

**Acceptance**:
- Given a developer has completed work on a feature branch, when they merge to `next`, then the changes are integrated with other pending features
- Given changes are merged to `next`, when CI/CD runs, then the staging environment reflects the combined changes

---

### US2 - Promote Next to Main (P1)

After validating that all features in `next` work correctly together, a developer or release manager promotes the `next` branch to main for production deployment.

**Acceptance**:
- Given `next` contains validated features, when promoted to main, then all integrated features are released together
- Given a promotion to main, when the merge completes, then `next` and main are synchronized

---

### US3 - Reset Next After Release (P2)

After promoting `next` to main, the `next` branch should be reset to match main, providing a clean slate for the next release cycle.

**Acceptance**:
- Given a successful promotion to main, when reset is performed, then `next` matches main exactly
- Given `next` is reset, when new features are merged, then they build upon the latest main

---

### US4 - Pre-release on Push to Next (P1)

Every push to the `next` branch automatically triggers a pre-release, producing CLI binaries, VS Code extension, and deploying the web app to `next.yapi.run`.

**Acceptance**:
- Given a push to `next`, when the workflow completes, then a GitHub pre-release is created with version `v<latest-tag>-next.<short-hash>`
- Given a pre-release, when viewing the release assets, then CLI binaries and VS Code extension `.vsix` are available
- Given a push to `next`, when Vercel deploys, then `next.yapi.run` reflects the latest changes

## Requirements

### Functional

- **FR-001**: The repository MUST have a protected `next` branch that serves as the staging/integration branch
- **FR-002**: Feature branches MUST be mergeable to `next` for integration testing
- **FR-003**: The `next` branch MUST be promotable to `main` when features are validated
- **FR-004**: The `next` branch MUST be resettable to match `main` after a release
- **FR-005**: Every push to `next` MUST trigger a pre-release workflow
- **FR-006**: Pre-releases MUST use version format `v<latest-stable-tag>-next.<7-char-commit-hash>`
- **FR-007**: Pre-releases MUST be marked as pre-release on GitHub (not "latest")
- **FR-008**: The web app MUST deploy to `next.yapi.run` on every push to `next`

### Branch Workflow

```
feature/* ─────┐
               │
feature/* ─────┼──▶ next ──▶ main
               │     │
feature/* ─────┘     │
                     │
              (reset after release)

On push to next:
  ├── CI runs (lint, test, build)
  ├── GoReleaser creates pre-release (v0.X.Y-next.<hash>)
  └── Vercel deploys to next.yapi.run
```

## Edge Cases

- What happens when conflicting features are both merged to `next`?
  - Conflicts must be resolved in the `next` branch before promotion
- What happens when a feature in `next` is found to be broken?
  - The feature branch owner must fix and re-merge, or the feature is reverted from `next`
- What happens when `next` diverges significantly from `main`?
  - Periodic rebasing or merging of `main` into `next` keeps them synchronized
- What happens when a pre-release fails?
  - The push is still accepted; release failures are surfaced via GitHub Actions status
- What happens when multiple pushes occur in quick succession?
  - Each push triggers its own release; GitHub handles concurrent workflows

## Success Criteria

- [ ] `next` branch exists and is protected from direct commits
- [ ] Feature branches can be merged to `next` without issues
- [ ] Changes in `next` can be promoted to `main` in a single operation
- [ ] After release, `next` can be reset to match `main`
- [ ] Push to `next` triggers pre-release workflow
- [ ] Pre-releases appear on GitHub with correct version format (`v0.X.Y-next.<hash>`)
- [ ] Pre-releases include CLI binaries and VS Code extension
- [ ] `next.yapi.run` serves the latest `next` branch web app
- [ ] Team members understand the workflow (documented in contributing guide)

## Assumptions

- The team uses a standard merge-based workflow (not rebase-only)
- Branch protection rules will be configured at the repository level
- Vercel domain configuration for `next.yapi.run` requires manual setup
- VS Code extension is included in next pre-releases (uploaded to same GitHub release)
- Homebrew is excluded from next pre-releases (GitHub pre-releases only)
