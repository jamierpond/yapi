# Feature Specification: Next Branch Workflow

**Branch**: `007-jp-next-branch` | **Created**: 2026-01-04 | **Status**: Draft

## Overview

Establish a staging branch named `next` that serves as an integration point between feature branches and the main production branch, enabling developers to validate changes before merging to main.

## User Stories

### US1 - Merge Feature to Next (P1)

A developer completes work on a feature branch and wants to validate it in a staging environment before merging to main. They merge their feature branch into `next`, which triggers any configured CI/CD pipelines for staging validation.

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

## Requirements

### Functional

- **FR-001**: The repository MUST have a protected `next` branch that serves as the staging/integration branch
- **FR-002**: Feature branches MUST be mergeable to `next` for integration testing
- **FR-003**: The `next` branch MUST be promotable to `main` when features are validated
- **FR-004**: The `next` branch MUST be resettable to match `main` after a release

### Branch Workflow

```
feature/* ─────┐
               │
feature/* ─────┼──▶ next ──▶ main
               │     │
feature/* ─────┘     │
                     │
              (reset after release)
```

## Edge Cases

- What happens when conflicting features are both merged to `next`?
  - Conflicts must be resolved in the `next` branch before promotion
- What happens when a feature in `next` is found to be broken?
  - The feature branch owner must fix and re-merge, or the feature is reverted from `next`
- What happens when `next` diverges significantly from `main`?
  - Periodic rebasing or merging of `main` into `next` keeps them synchronized

## Success Criteria

- [ ] `next` branch exists and is protected from direct commits
- [ ] Feature branches can be merged to `next` without issues
- [ ] Changes in `next` can be promoted to `main` in a single operation
- [ ] After release, `next` can be reset to match `main`
- [ ] Team members understand the workflow (documented in contributing guide)

## Assumptions

- The team uses a standard merge-based workflow (not rebase-only)
- CI/CD pipelines will be configured separately to run on the `next` branch
- Branch protection rules will be configured at the repository level
- This feature focuses on establishing the branch structure; automation tooling is out of scope
