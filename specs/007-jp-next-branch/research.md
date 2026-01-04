# Research: Next Branch Workflow with Nightly Releases

## Current State Analysis

### GitHub Workflows

| Workflow | Current Branches | Trigger Type | Needs Update |
|----------|------------------|--------------|--------------|
| cli.yml | `main` only | push/PR | Yes - add `next` |
| codecov.yml | `main` only | push/PR | Yes - add `next` |
| github-action-dist.yml | any (path-based) | push/PR | No |
| installer-tests.yml | `main` only | push/PR | Yes - add `next` |
| release.yaml | tags only (`v*`) | push | No - keep for stable |
| vscode-extension-build.yml | `main` only | push/PR | Yes - add `next` |
| vscode-extension-publish.yml | tags only (`v*`) | push | No - stable only |
| web-tests.yml | `main` only | push/PR | Yes - add `next` |

### GoReleaser Configuration

**Current** (`.goreleaser.yaml`):
- Uses `homebrew_casks` (not `brews`) for Homebrew
- Publishes to `jamierpond/homebrew-yapi` repository
- AUR package: `yapi-bin`
- Version template uses `{{.Version}}` from git tags

**Key insight**: GoReleaser can run in "snapshot" mode for non-tag builds, but this doesn't create GitHub releases. For nightly, we need a different approach.

### Release Process

**Current release flow** (from `cli/scripts/bump.sh` and `Makefile`):
1. `bump.sh` only allows releases from `main` or `develop`
2. Tags are created with format `v*.*.*`
3. Release workflow triggers on tag push
4. GoReleaser handles distribution
5. VS Code extension publishes on tag push

### Vercel Configuration

**Current** (`vercel.json`):
- Framework: Next.js
- Build command: `bash ./cli/scripts/vercel-build.sh`
- Output: `./apps/web/.next`

Vercel branch deployments are configured via Vercel dashboard. For `nightly.yapi.run`:
- Configure custom domain alias for `next` branch in Vercel dashboard
- Branch: `next` → Domain: `nightly.yapi.run`

---

## Decisions

### 1. Nightly Release Workflow Strategy

**Decision**: Create a new `nightly.yaml` workflow separate from `release.yaml`

**Rationale**:
- Stable releases use tags (`v*`) and create "latest" releases
- Nightly releases trigger on push to `next` and create pre-releases
- Different GoReleaser configurations needed (snapshot mode vs full release)
- Keeps stable release workflow unchanged (minimal code principle: don't modify working code)

**Alternatives considered**:
- Single workflow with conditionals: Rejected - adds complexity to stable release path
- Modify release.yaml: Rejected - risk breaking stable releases

### 2. Nightly Version Scheme

**Decision**: `v<latest-tag>-nightly.<short-hash>`

**Implementation**:
```bash
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
SHORT_HASH=$(git rev-parse --short HEAD)
NIGHTLY_VERSION="${LATEST_TAG}-nightly.${SHORT_HASH}"
# Example: v0.5.2-nightly.abc1234
```

**Rationale**:
- Semver compliant (pre-release suffix)
- Sorts correctly (v0.5.2-nightly.xxx < v0.5.3)
- Short hash provides traceability
- Base version shows what stable it's built on

### 3. GoReleaser Nightly Configuration

**Decision**: Create `.goreleaser.nightly.yaml` for nightly builds

**Key differences from stable**:
- Uses `--snapshot` flag (no tag required)
- Creates pre-release on GitHub
- Updates `yapi-nightly` Homebrew cask (not `yapi`)
- Skips AUR (nightly users can build from source)
- Version override via environment variable

**Configuration approach**:
```yaml
# .goreleaser.nightly.yaml
version: 2
project_name: yapi-nightly

# Override version from environment
snapshot:
  version_template: "{{ .Env.NIGHTLY_VERSION }}"

release:
  prerelease: true
  name_template: "Nightly {{ .Env.NIGHTLY_VERSION }}"

homebrew_casks:
  - name: yapi-nightly
    # ... same structure but different cask name
```

### 4. Homebrew Nightly Cask

**Decision**: Separate `yapi-nightly` cask in same tap

**Implementation**:
- Cask name: `yapi-nightly`
- Install command: `brew install yapi/tap/yapi-nightly`
- Same tap repository: `jamierpond/homebrew-yapi`
- Different cask file: `Casks/yapi-nightly.rb`

**Rationale**:
- Users can have both stable and nightly installed
- Clear distinction between channels
- Single tap to manage

### 5. Vercel Nightly Subdomain

**Decision**: Configure `nightly.yapi.run` in Vercel dashboard

**Steps**:
1. Add custom domain `nightly.yapi.run` to Vercel project
2. Configure branch alias: `next` branch → `nightly.yapi.run`
3. DNS: Add CNAME record for `nightly.yapi.run` → `cname.vercel-dns.com`

**Note**: This is a manual configuration step, not code.

### 6. Workflow Branch Configuration

**Decision**: Add `next` to branch triggers alongside `main`

**Rationale**:
- CI should run on `next` to validate integrations before promotion to main
- All existing workflows that run on `main` should also run on `next`
- Tag-based workflows (release, vscode-publish) remain unchanged

---

## Files to Modify/Create

| File | Action | Purpose |
|------|--------|---------|
| `.github/workflows/nightly.yaml` | Create | New nightly release workflow |
| `.goreleaser.nightly.yaml` | Create | GoReleaser config for nightlies |
| `.github/workflows/cli.yml` | Modify | Add `next` to branches |
| `.github/workflows/codecov.yml` | Modify | Add `next` to branches |
| `.github/workflows/installer-tests.yml` | Modify | Add `next` to branches |
| `.github/workflows/vscode-extension-build.yml` | Modify | Add `next` to branches |
| `.github/workflows/web-tests.yml` | Modify | Add `next` to branches |
| `cli/scripts/bump.sh` | Modify | Add `next` to allowed branches |
| `Makefile` | Modify | Add `next` to release target |

---

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

---

## Out of Scope

- VS Code extension nightly (stable only per clarification)
- AUR nightly package (users can build from source)
- Automatic cleanup of old nightly releases
