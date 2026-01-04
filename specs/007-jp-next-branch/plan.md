# Implementation Plan: Next Branch Workflow with Pre-releases

**Branch**: `007-jp-next-branch` | **Date**: 2026-01-04 | **Spec**: [spec.md](./spec.md)

## Summary

Establish a `next` branch as an unstable integration point with automatic pre-releases on every push. This includes:
1. CI workflows running on `next` branch
2. New pre-release workflow creating GitHub pre-releases (`v0.X.Y-next.<hash>`)
3. `next.yapi.run` subdomain for web app

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
| Simplicity | [x] | GitHub pre-releases only, no package manager updates |
| Dogfooding | [x] | `next.yapi.run` serves as dogfooding environment |
| Minimal Code | [x] | Reuses existing GoReleaser build config |

## Affected Areas

```text
.github/workflows/
├── next-release.yaml          # NEW - pre-release workflow (CLI + VS Code extension)
├── cli.yml                    # Add 'next' to branches
├── codecov.yml                # Add 'next' to branches
├── installer-tests.yml        # Add 'next' to branches
├── vscode-extension-build.yml # Add 'next' to branches
└── web-tests.yml              # Add 'next' to branches

.goreleaser.next.yaml          # NEW - GoReleaser config for next pre-releases

cli/scripts/
└── bump.sh                    # Add 'next' to allowed branches

Makefile                       # Add 'next' to release target
```

## Implementation Approach

1. **Create next release workflow** (`.github/workflows/next-release.yaml`):
   - Trigger on push to `next` branch
   - Calculate version: `v<latest-tag>-next.<short-hash>`
   - Run GoReleaser with next config (CLI binaries)
   - Build VS Code extension and upload to same release
   - Create GitHub pre-release

2. **Create next GoReleaser config** (`.goreleaser.next.yaml`):
   - Same build configuration as stable
   - Pre-release mode enabled
   - No Homebrew/AUR updates (users download directly from GitHub)

3. **Update CI workflows** to run on `next` branch:
   - Add `next` to push/PR triggers in 5 workflow files
   - Same CI validation as `main` branch

4. **Update release scripts** to allow `next`:
   - `cli/scripts/bump.sh`: add `next` to allowed branches
   - `Makefile`: add `next` to release target

5. **Manual configuration** (post-merge):
   - Vercel: Configure `next.yapi.run` subdomain for `next` branch
   - DNS: Add CNAME record for `next.yapi.run`
   - GitHub: Branch protection rules for `next`

## Detailed Changes

### 1. New File: `.github/workflows/next-release.yaml`

```yaml
name: Next Release

on:
  push:
    branches: [next]

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
          submodules: recursive

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      - uses: pnpm/action-setup@v4
        with:
          version: 9

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'pnpm'

      - name: Calculate version
        id: version
        run: |
          LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
          SHORT_HASH=$(git rev-parse --short HEAD)
          echo "version=${LATEST_TAG}-next.${SHORT_HASH}" >> $GITHUB_OUTPUT

      # Build CLI with GoReleaser
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean --config .goreleaser.next.yaml
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GORELEASER_CURRENT_TAG: ${{ steps.version.outputs.version }}
          POSTHOG_API_KEY: ${{ secrets.POSTHOG_API_KEY }}
          POSTHOG_API_HOST: https://us.i.posthog.com

      # Build VS Code extension
      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Build VS Code extension
        run: pnpm package:extension

      # Upload VS Code extension to the same release
      - name: Upload VS Code extension to release
        run: |
          VSIX_FILE=$(ls integrations/vscode/*.vsix | head -1)
          gh release upload "${{ steps.version.outputs.version }}" "$VSIX_FILE" --clobber
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 2. New File: `.goreleaser.next.yaml`

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
  name_template: "Next {{ .Version }}"

changelog:
  disable: true
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
2. Add `next.yapi.run` as custom domain
3. Go to Project Settings > Git
4. Under "Branch Aliases", map `next` → `next.yapi.run`

### DNS (wherever yapi.run is hosted)
1. Add CNAME record: `next` → `cname.vercel-dns.com`

### GitHub Repository Settings
1. Go to Settings > Branches
2. Add branch protection rule for `next`

## Verification

After implementation:
- [ ] Push to `next` triggers all CI workflows
- [ ] Push to `next` triggers next-release workflow
- [ ] Pre-releases appear on GitHub with correct version format (`v0.X.Y-next.<hash>`)
- [ ] Pre-releases include CLI binaries for all platforms
- [ ] Pre-releases include VS Code extension `.vsix` file
- [ ] `next.yapi.run` serves the latest `next` branch web app
- [ ] PR to `next` triggers all CI workflows
- [ ] `make release` works from `next` branch
- [ ] `bump.sh` works from `next` branch
