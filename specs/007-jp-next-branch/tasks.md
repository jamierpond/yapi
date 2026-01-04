# Tasks: Next Branch Workflow

**Input**: `specs/007-jp-next-branch/plan.md`

## Format

- `[P]` = Can run in parallel
- `[US#]` = User story reference
- Include file paths: `.github/workflows/cli.yml`

---

## Phase 1: Setup

> Goal: Create the `next` branch in the repository

- [ ] T001 Create `next` branch from `main` (manual git operation)

**Checkpoint**: `next` branch exists in remote repository

---

## Phase 2: User Story 1 - Merge Feature to Next (P1)

> Goal: CI runs when features are merged to `next`
>
> Independent Test: Push to `next` triggers all CI workflows

- [ ] T002 [P] [US1] Add `next` to push/PR branches in `.github/workflows/cli.yml`
- [ ] T003 [P] [US1] Add `next` to push/PR branches in `.github/workflows/codecov.yml`
- [ ] T004 [P] [US1] Add `next` to push/PR branches in `.github/workflows/installer-tests.yml`
- [ ] T005 [P] [US1] Add `next` to push/PR branches in `.github/workflows/vscode-extension-build.yml`
- [ ] T006 [P] [US1] Add `next` to push/PR branches in `.github/workflows/web-tests.yml`

**Checkpoint**: Push to `next` triggers all 5 CI workflows

---

## Phase 3: User Story 2 - Promote Next to Main (P1)

> Goal: Release scripts allow operations from `next` branch
>
> Independent Test: `make release` and `bump.sh` work from `next` branch

- [ ] T007 [P] [US2] Add `next` to allowed branches in `cli/scripts/bump.sh`
- [ ] T008 [P] [US2] Add `next` to release target branches in `Makefile`

**Checkpoint**: `./cli/scripts/bump.sh` runs without error on `next` branch

---

## Phase 4: User Story 3 - Reset Next After Release (P2)

> Goal: Document the reset workflow for post-release cleanup
>
> Independent Test: N/A - process documentation only

- [ ] T009 [US3] Document reset workflow in CONTRIBUTING.md or development docs

**Checkpoint**: Reset process is documented and accessible to team

---

## Phase 5: Manual Configuration (Post-Merge)

> Goal: Complete external platform configurations
>
> Note: These are manual steps performed in web dashboards after code changes are merged

- [ ] T010 Configure `next` as preview branch in Vercel dashboard (Settings > Git)
- [ ] T011 Add branch protection rules for `next` in GitHub repository settings

**Checkpoint**: All verification items from plan.md pass

---

## Dependencies

```text
T001 (Setup)
  │
  ├──▶ T002-T006 [Parallel - US1 CI Workflows]
  │         │
  │         └──▶ T010 (Vercel - requires workflows in place)
  │
  └──▶ T007-T008 [Parallel - US2 Release Scripts]
            │
            └──▶ T009 (US3 Documentation)
                      │
                      └──▶ T011 (GitHub Protection - final step)
```

## Parallel Execution

**Maximum parallelism** (after T001):
- T002, T003, T004, T005, T006, T007, T008 can all run in parallel

**Recommended batches**:
1. T001 (branch creation)
2. T002-T008 in parallel (all code changes)
3. T009 (documentation)
4. T010, T011 in parallel (external configuration)

---

## Verification

```bash
# After T002-T006: Push test commit to next
# Verify all workflows trigger in GitHub Actions

# After T007-T008: On next branch
./cli/scripts/bump.sh patch  # Should not error (dry run verification)

# Full verification checklist from plan.md:
# - [ ] Push to `next` triggers all CI workflows
# - [ ] PR to `next` triggers all CI workflows
# - [ ] `make release` works from `next` branch
# - [ ] `bump.sh` works from `next` branch
# - [ ] Vercel deploys preview for `next` branch
```

---

## Notes

- All workflow changes are additive (adding `next` to existing arrays)
- No new workflow logic is introduced
- Tag-based workflows (release.yaml, vscode-extension-publish.yml) remain unchanged
- Manual configuration steps (T010, T011) require repository admin access
