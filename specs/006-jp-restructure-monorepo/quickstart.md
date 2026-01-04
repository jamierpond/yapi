# Quickstart: Monorepo Restructure

**Feature**: 006-jp-restructure-monorepo

## Overview

This restructure moves files/directories and updates configuration. No new code is written.

## Implementation Order

Execute in this exact order to minimize broken states:

### Step 1: Create Target Directories

```bash
mkdir -p cli integrations assets
```

### Step 2: Move Go CLI (all at once)

```bash
mv cmd cli/
mv internal cli/
mv scripts cli/
mv bin cli/
mv go.mod cli/
mv go.sum cli/
```

### Step 3: Move Integrations

```bash
mv action integrations/github-action
mv extensions/vscode-extension integrations/vscode
mv lua integrations/nvim
rmdir extensions  # Should be empty now
```

### Step 4: Hoist Packages

```bash
mv apps/vscode-webview packages/vscode-ui
# For git submodule, update .gitmodules first, then:
mv apps/web/madea-blog-core packages/madea-blog-core
```

### Step 5: Move Assets

```bash
mv pup.jpg assets/
mv icon.svg assets/
mv og-image.png assets/
```

### Step 6: Update Configuration Files

**pnpm-workspace.yaml**:
```yaml
packages:
  - 'apps/*'
  - 'packages/*'
  - 'integrations/*'
```

**.gitmodules**:
```gitmodules
[submodule "packages/madea-blog-core"]
    path = packages/madea-blog-core
    url = https://github.com/jamierpond/madea-blog-core.git
```

**Makefile** (build target):
```makefile
build:
    cd cli && go build -ldflags "$(LDFLAGS)" -o ./bin/yapi ./cmd/yapi
```

**integrations/vscode/package.json** (copy:webview script):
```json
"copy:webview": "rm -rf media && mkdir -p media && cp -R ../../packages/vscode-ui/dist/. media/"
```

### Step 7: Verify

```bash
# Go CLI
cd cli && go build ./cmd/yapi && go test ./... && cd ..

# TypeScript packages
pnpm install
pnpm -r build

# VS Code extension specifically
pnpm --filter yapi-extension build
```

## Key Files Changed

| File | Change |
|------|--------|
| `Makefile` | Update paths to `cli/` |
| `pnpm-workspace.yaml` | Add `integrations/*` |
| `.gitmodules` | Update submodule path |
| `integrations/vscode/package.json` | Update copy:webview path |

## Validation Checklist

- [ ] `make build` succeeds
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] `pnpm install` succeeds
- [ ] VS Code extension builds
- [ ] Web app builds
- [ ] No broken imports in any package
