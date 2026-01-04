# Tasks: Next Branch Workflow with Pre-releases

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

## Phase 4: User Story 4 - Pre-release on Push to Next (P1)

> Goal: Every push to `next` creates a GitHub pre-release with CLI + VS Code extension
>
> Independent Test: Push to `next` creates pre-release with correct version and assets

- [ ] T009 [P] [US4] Create `.goreleaser.next.yaml` (copy from `.goreleaser.yaml`, set prerelease: true, remove homebrew_casks/aurs)
- [ ] T010 [US4] Create `.github/workflows/next-release.yaml` with:
  - Trigger on push to `next`
  - Calculate version `v<latest-tag>-next.<short-hash>`
  - Run GoReleaser with `.goreleaser.next.yaml`
  - Build VS Code extension with `pnpm package:extension`
  - Upload `.vsix` to same release with `gh release upload`

**Checkpoint**: Push to `next` creates pre-release with CLI binaries and VS Code extension

---

## Phase 5: User Story 3 - Reset Next After Release (P2)

> Goal: Document the reset workflow for post-release cleanup
>
> Independent Test: N/A - process documentation only

- [ ] T011 [US3] Document reset workflow in CONTRIBUTING.md or development docs

**Checkpoint**: Reset process is documented and accessible to team

---

## Phase 6: Manual Configuration (Post-Merge)

> Goal: Complete external platform configurations
>
> Note: These are manual steps performed in web dashboards after code changes are merged

- [ ] T012 Add `next.yapi.run` custom domain in Vercel dashboard (Settings > Domains)
- [ ] T013 Configure branch alias: `next` → `next.yapi.run` in Vercel (Settings > Git)
- [ ] T014 Add DNS CNAME record: `next` → `cname.vercel-dns.com`
- [ ] T015 Add branch protection rules for `next` in GitHub repository settings

**Checkpoint**: All verification items from plan.md pass

---

## Dependencies

```text
T001 (Setup)
  │
  ├──▶ T002-T006 [Parallel - US1 CI Workflows]
  │
  ├──▶ T007-T008 [Parallel - US2 Release Scripts]
  │
  └──▶ T009-T010 [US4 Pre-release Workflow]
            │
            └──▶ T012-T014 (Vercel/DNS - requires workflow)
                      │
                      └──▶ T011 (US3 Documentation)
                                │
                                └──▶ T015 (GitHub Protection - final step)
```

## Parallel Execution

**Maximum parallelism** (after T001):
- T002, T003, T004, T005, T006, T007, T008, T009 can all run in parallel

**Recommended batches**:
1. T001 (branch creation)
2. T002-T009 in parallel (all config file changes)
3. T010 (workflow file - depends on T009 goreleaser config)
4. T011 (documentation)
5. T012-T015 in parallel (external configuration)

---

## Verification

```bash
# After T002-T006: Push test commit to next
# Verify all CI workflows trigger in GitHub Actions

# After T007-T008: On next branch
./cli/scripts/bump.sh patch  # Should not error (dry run verification)

# After T009-T010: Push to next
# Verify pre-release created with:
# - Version format: v0.X.Y-next.<hash>
# - CLI binaries for all platforms
# - VS Code extension .vsix file

# Full verification checklist from plan.md:
# - [ ] Push to `next` triggers all CI workflows
# - [ ] Push to `next` triggers next-release workflow
# - [ ] Pre-releases appear with correct version format
# - [ ] Pre-releases include CLI binaries and VS Code extension
# - [ ] `next.yapi.run` serves the latest `next` branch web app
# - [ ] PR to `next` triggers all CI workflows
# - [ ] `make release` works from `next` branch
# - [ ] `bump.sh` works from `next` branch
```

---

## Notes

- CI workflow changes are additive (adding `next` to existing arrays)
- New `next-release.yaml` workflow handles pre-releases separately from stable `release.yaml`
- `.goreleaser.next.yaml` is a simplified copy without Homebrew/AUR
- VS Code extension is built and uploaded to same GitHub release as CLI
- Manual configuration steps (T012-T015) require repository admin access
