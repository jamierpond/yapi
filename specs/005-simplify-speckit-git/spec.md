# Feature Specification: Simplify Speckit Git Handling

**Branch**: `005-simplify-speckit-git` | **Created**: 2026-01-03 | **Status**: Draft

## Overview

Remove all git branch creation and management from speckit. Speckit should only validate that the user is on a valid feature branch (not a protected branch) and proceed with spec/plan/task workflows without manipulating git state.

## User Stories

### US1 - Run Speckit on Existing Branch (P1)

As a developer, I want to run `/speckit.specify` on my current branch without speckit trying to create or switch branches, so that I maintain full control over my git workflow.

**CLI Usage**:
```bash
# User is already on their feature branch
claude "/speckit.specify add user authentication"
```

**Acceptance**:
- Given I am on branch `feature/my-feature`, when I run `/speckit.specify`, then speckit validates my branch and proceeds without any git operations
- Given I am on branch `main`, when I run `/speckit.specify`, then speckit rejects with "Cannot run on protected branch"
- Given I am on branch `master`, when I run `/speckit.specify`, then speckit rejects with "Cannot run on protected branch"

---

### US2 - Valid Feature Branch Detection (P1)

As a developer, I want speckit to recognize various feature branch naming conventions, so that I can use my preferred branching strategy.

**Acceptance**:
- Given I am on `feature/foo`, when I run speckit, then it proceeds
- Given I am on `jp/some-feature`, when I run speckit, then it proceeds (initials prefix)
- Given I am on `123-feature-name`, when I run speckit, then it proceeds (numbered branches)
- Given I am on `fix/bug-description`, when I run speckit, then it proceeds

## Requirements

### Functional

- **FR-001**: Speckit MUST NOT create, checkout, or switch git branches
- **FR-002**: Speckit MUST NOT run `git fetch`, `git pull`, or any remote operations
- **FR-003**: Speckit MUST validate current branch is not in protected list (`main`, `master`, `develop`, `release/*`)
- **FR-004**: Speckit MUST derive feature directory name from current branch name
- **FR-005**: Speckit MUST create spec files in `specs/<branch-name>/` directory
- **FR-006**: Speckit MUST proceed if branch validation passes, without requiring specific naming conventions beyond "not protected"

### Protected Branches

| Branch Pattern | Protected | Reason |
|----------------|-----------|--------|
| `main`         | Yes       | Primary branch |
| `master`       | Yes       | Legacy primary |
| `develop`      | Yes       | Integration branch |
| `release/*`    | Yes       | Release branches |
| Everything else | No       | Feature work |

## Edge Cases

- What happens when user is in detached HEAD state? **Reject with clear error message**
- What happens when current directory is not a git repo? **Reject with clear error message**
- What happens when branch name contains special characters? **Sanitize for directory name (replace `/` with `-`)**

## Success Criteria

- [ ] No git branch creation or switching occurs during any speckit command
- [ ] Protected branch detection works correctly
- [ ] Spec files are created in correct directory based on current branch
- [ ] Users maintain full control of their git workflow
- [ ] Existing speckit functionality (spec/plan/task generation) works unchanged
- [ ] Error messages clearly explain why a branch is rejected

## Assumptions

- Users are responsible for creating and switching to their feature branches before running speckit
- The current branch name is sufficient to derive a meaningful spec directory name
- Users prefer explicit git control over automated branch management
