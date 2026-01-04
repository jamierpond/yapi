# Feature Specification: Next Branch Workflow

**Branch**: `007-jp-next-branch` | **Created**: 2026-01-04 | **Status**: Draft

## Overview

Establish a staging branch named `next` that serves as an unstable/nightly integration point between feature branches and the main production branch. The `next` branch enables developers to validate changes before merging to main, and automatically produces nightly releases on every push.

## Clarifications

### Session 2026-01-04

- Q: When should nightly releases trigger? → A: Every push to `next`
- Q: What version format for nightlies? → A: `v0.X.Y-nightly.<short-hash>` (semver compliant, 7-char hash)
- Q: What subdomain for nightly web app? → A: Single `nightly.yapi.run` (always latest push)
- Q: What artifacts for nightly releases? → A: Full GoReleaser with `yapi-nightly` Homebrew formula, marked as pre-release
- Q: How is nightly base version determined? → A: Latest stable tag (e.g., v0.5.2-nightly.abc1234)
- Q: VS Code extension nightly releases? → A: Skipped - extension only releases with stable

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

### US4 - Nightly Release on Push (P1)

Every push to the `next` branch automatically triggers a nightly release, producing CLI binaries and deploying the web app to the nightly subdomain.

**Acceptance**:
- Given a push to `next`, when the workflow completes, then a GitHub pre-release is created with version `v<latest-tag>-nightly.<short-hash>`
- Given a nightly release, when a user installs via Homebrew, then `brew install yapi/tap/yapi-nightly` provides the latest nightly binary
- Given a push to `next`, when Vercel deploys, then `nightly.yapi.run` reflects the latest changes

## Requirements

### Functional

- **FR-001**: The repository MUST have a protected `next` branch that serves as the staging/integration branch
- **FR-002**: Feature branches MUST be mergeable to `next` for integration testing
- **FR-003**: The `next` branch MUST be promotable to `main` when features are validated
- **FR-004**: The `next` branch MUST be resettable to match `main` after a release
- **FR-005**: Every push to `next` MUST trigger a nightly release workflow
- **FR-006**: Nightly releases MUST use version format `v<latest-stable-tag>-nightly.<7-char-commit-hash>`
- **FR-007**: Nightly releases MUST be marked as pre-release on GitHub (not "latest")
- **FR-008**: Nightly CLI binaries MUST be installable via `brew install yapi/tap/yapi-nightly`
- **FR-009**: The web app MUST deploy to `nightly.yapi.run` on every push to `next`

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
  ├── GoReleaser creates pre-release (v0.X.Y-nightly.<hash>)
  ├── Homebrew yapi-nightly formula updated
  └── Vercel deploys to nightly.yapi.run
```

## Edge Cases

- What happens when conflicting features are both merged to `next`?
  - Conflicts must be resolved in the `next` branch before promotion
- What happens when a feature in `next` is found to be broken?
  - The feature branch owner must fix and re-merge, or the feature is reverted from `next`
- What happens when `next` diverges significantly from `main`?
  - Periodic rebasing or merging of `main` into `next` keeps them synchronized
- What happens when a nightly release fails?
  - The push is still accepted; release failures are surfaced via GitHub Actions status
- What happens when multiple pushes occur in quick succession?
  - Each push triggers its own release; GitHub handles concurrent workflows

## Success Criteria

- [ ] `next` branch exists and is protected from direct commits
- [ ] Feature branches can be merged to `next` without issues
- [ ] Changes in `next` can be promoted to `main` in a single operation
- [ ] After release, `next` can be reset to match `main`
- [ ] Push to `next` triggers nightly release workflow
- [ ] Nightly releases appear as pre-releases on GitHub with correct version format
- [ ] `brew install yapi/tap/yapi-nightly` installs the latest nightly CLI
- [ ] `nightly.yapi.run` serves the latest `next` branch web app
- [ ] Team members understand the workflow (documented in contributing guide)

## Assumptions

- The team uses a standard merge-based workflow (not rebase-only)
- Branch protection rules will be configured at the repository level
- Vercel domain configuration for `nightly.yapi.run` requires manual setup
- Homebrew tap repository exists and accepts automated formula updates
- VS Code extension is excluded from nightly releases (stable only)
