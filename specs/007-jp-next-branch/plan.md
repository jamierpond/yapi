# Implementation Plan: Next Branch Workflow with Nightly Releases

**Branch**: `007-jp-next-branch` | **Date**: 2026-01-04 | **Spec**: [spec.md](./spec.md)

## Summary

Establish a `next` branch as an unstable/nightly integration point with automatic releases on every push. This includes:
1. CI workflows running on `next` branch
2. New nightly release workflow with GoReleaser
3. Separate `yapi-nightly` Homebrew cask
4. `nightly.yapi.run` subdomain for web app

## Technical Context

**Type**: Infrastructure/DevOps - no application code changes
**Files**: GitHub workflow YAML, GoReleaser config, shell scripts, Makefile
**Testing**: Manual verification of workflow triggers and release artifacts
**Build**: N/A (configuration changes only)

## Constitution Check

*Must pass before implementation.*

| Principle | Status | Notes |
|-----------|--------|-------|
| CLI-First | [x] | N/A - infrastructure only, no CLI changes |
| Git-Friendly | [x] | All configs in plain-text YAML files |
| Protocol Agnostic | [x] | N/A - infrastructure only |
| Simplicity | [x] | Separate nightly workflow avoids modifying stable release path |
| Dogfooding | [x] | `next.yapi.run` serves as dogfooding environment |
| Minimal Code | [x] | New workflow reuses existing GoReleaser patterns |

## Affected Areas

```text
.github/workflows/
├── nightly.yaml               # NEW - nightly release workflow
├── cli.yml                    # Add 'next' to branches
├── codecov.yml                # Add 'next' to branches
├── installer-tests.yml        # Add 'next' to branches
├── vscode-extension-build.yml # Add 'next' to branches
└── web-tests.yml              # Add 'next' to branches

.goreleaser.nightly.yaml       # NEW - GoReleaser config for nightlies

cli/scripts/
└── bump.sh                    # Add 'next' to allowed branches

Makefile                       # Add 'next' to release target
```

## Implementation Approach

1. **Create nightly release workflow** (`.github/workflows/nightly.yaml`):
   - Trigger on push to `next` branch
   - Calculate version: `v<latest-tag>-nightly.<short-hash>`
   - Run GoReleaser with nightly config
   - Create GitHub pre-release

2. **Create nightly GoReleaser config** (`.goreleaser.nightly.yaml`):
   - Same build configuration as stable
   - Pre-release mode enabled
   - Publishes `yapi-nightly` Homebrew cask
   - Skips AUR (nightly users can build from source)

3. **Update CI workflows** to run on `next` branch:
   - Add `next` to push/PR triggers in 5 workflow files
   - Same CI validation as `main` branch

4. **Update release scripts** to allow `next`:
   - `cli/scripts/bump.sh`: add `next` to allowed branches
   - `Makefile`: add `next` to release target

5. **Manual configuration** (post-merge):
   - Vercel: Configure `nightly.yapi.run` subdomain for `next` branch
   - DNS: Add CNAME record for `nightly.yapi.run`
   - GitHub: Branch protection rules for `next`

## Detailed Changes

### 1. New File: `.github/workflows/nightly.yaml`

```yaml
name: Nightly Release

on:
  push:
    branches: [next]

permissions:
  contents: write

jobs:
  nightly:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Calculate nightly version
        id: version
        run: |
          LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
          SHORT_HASH=$(git rev-parse --short HEAD)
          echo "version=${LATEST_TAG}-nightly.${SHORT_HASH}" >> $GITHUB_OUTPUT

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean --config .goreleaser.nightly.yaml
        env:
          GITHUB_TOKEN: ${{ secrets.HOMEBREW_GITHUB_PAT }}
          GORELEASER_CURRENT_TAG: ${{ steps.version.outputs.version }}
          POSTHOG_API_KEY: ${{ secrets.POSTHOG_API_KEY }}
          POSTHOG_API_HOST: https://us.i.posthog.com
```

### 2. New File: `.goreleaser.nightly.yaml`

```yaml
version: 2
project_name: yapi

builds:
  - id: yapi
    dir: cli
    main: ./cmd/yapi
    binary: yapi
    ldflags:
      - -s -w
      - -X main.version={{ .Version }}
      - -X main.commit={{ .ShortCommit }}
      - -X main.date={{ .Date }}
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]

archives:
  - id: yapi_archive
    ids: [yapi]
    formats: [tar.gz]
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"
    files:
      - LICENSE
    format_overrides:
      - goos: windows
        formats: [zip]

checksum:
  name_template: checksums.txt

release:
  prerelease: true
  name_template: "Nightly {{ .Version }}"

changelog:
  disable: true

homebrew_casks:
  - name: yapi-nightly
    description: "CLI-first, offline-first, git-friendly API client (nightly build)"
    homepage: "https://github.com/jamierpond/yapi"
    binaries:
      - yapi
    directory: Casks
    hooks:
      post:
        install: |
          system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/yapi"]
    caveats: |
      This is a NIGHTLY build - unstable and may contain bugs.
      For stable releases, use: brew install yapi/tap/yapi
    repository:
      owner: jamierpond
      name: homebrew-yapi
      branch: main
```

### 3. CI Workflow Updates (5 files)

Add `next` to branch triggers:

**cli.yml**, **codecov.yml**, **installer-tests.yml**, **vscode-extension-build.yml**, **web-tests.yml**:
```yaml
on:
  push:
    branches: [ main, next ]
  pull_request:
    branches: [ main, next ]
```

### 4. cli/scripts/bump.sh (lines 8-9)

```bash
if [[ "$CURRENT_BRANCH" != "main" && "$CURRENT_BRANCH" != "develop" && "$CURRENT_BRANCH" != "next" ]]; then
    echo "Error: Releases can only be made from 'main', 'develop', or 'next' branches"
```

### 5. Makefile (lines 105-106)

```makefile
if [ "$$BRANCH" != "main" ] && [ "$$BRANCH" != "develop" ] && [ "$$BRANCH" != "next" ]; then \
    echo "Error: Releases can only be made from 'main', 'develop', or 'next' branches"; \
```

## Manual Configuration Steps

### Vercel Dashboard
1. Go to Project Settings > Domains
2. Add `nightly.yapi.run` as custom domain
3. Go to Project Settings > Git
4. Under "Branch Aliases", map `next` → `nightly.yapi.run`

### DNS (wherever yapi.run is hosted)
1. Add CNAME record: `nightly` → `cname.vercel-dns.com`

### GitHub Repository Settings
1. Go to Settings > Branches
2. Add branch protection rule for `next`

## Verification

After implementation:
- [ ] Push to `next` triggers all CI workflows
- [ ] Push to `next` triggers nightly release workflow
- [ ] Nightly releases appear as pre-releases on GitHub
- [ ] Version format is correct: `v0.X.Y-nightly.<hash>`
- [ ] `brew install yapi/tap/yapi-nightly` installs nightly binary
- [ ] `nightly.yapi.run` serves the latest `next` branch web app
- [ ] PR to `next` triggers all CI workflows
- [ ] `make release` works from `next` branch
- [ ] `bump.sh` works from `next` branch
