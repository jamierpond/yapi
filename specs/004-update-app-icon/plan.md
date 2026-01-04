# Implementation Plan: Update App Icon

**Branch**: `001-update-app-icon` | **Date**: 2026-01-03 | **Spec**: [spec.md](./spec.md)

## Summary

Create a dynamic icon generation system with a shared implementation between the favicon (`icon.tsx`) and a new API endpoint (`/api/icon/[size]`). The icon will be 256x256 by default, use JetBrains Mono Bold font, and incorporate gradient styling from the landing page.

## Technical Context

**Framework**: Next.js 14+ (App Router)
**Key Packages**: `next/og` (ImageResponse), existing `COLORS` constants
**Font**: JetBrains Mono Bold (already in `public/fonts/`)
**Testing**: yapi test file for API validation
**Build**: `pnpm build` in apps/web

## Constitution Check

*Must pass before implementation.*

| Principle | Status | Notes |
|-----------|--------|-------|
| CLI-First | [x] | N/A - webapp feature, but testable via yapi CLI |
| Git-Friendly | [x] | All config in code, no binary formats |
| Protocol Agnostic | [x] | N/A - webapp-only feature |
| Simplicity | [x] | Single shared function, minimal abstraction |
| Dogfooding | [x] | API testable via yapi test file |

## Affected Areas

```text
apps/web/app/
├── icon.tsx                    # UPDATE: Use shared generateIcon()
├── lib/
│   └── icon.ts                 # NEW: Shared icon generation logic
└── api/
    └── icon/
        └── [size]/
            └── route.ts        # NEW: Dynamic icon API endpoint
```

## Implementation Approach

1. **Create shared icon generator** (`apps/web/app/lib/icon.ts`):
   - Export `generateIcon(size: number)` function returning `ImageResponse`
   - Load JetBrains Mono Bold font
   - Render "y" character with gradient background inspired by landing page
   - Scale font size proportionally to icon size

2. **Update existing favicon** (`apps/web/app/icon.tsx`):
   - Import and use `generateIcon(256)` from shared module
   - Update export size to 256x256

3. **Create dynamic API route** (`apps/web/app/api/icon/[size]/route.ts`):
   - Parse and validate size parameter (16, 32, 64, 128, 256)
   - Return 400 for invalid sizes
   - Add cache headers for performance
   - Call `generateIcon(size)` and return response

4. **Create yapi test file** (`apps/web/icon-api.yapi.yml`):
   - Chain requests for each valid size with `output_file`
   - Include invalid size test case

## Visual Design Implementation

The gradient background will use:
- Base: `COLORS.bg` (#0a0a0a)
- Gradient overlay: radial gradient from `COLORS.accent` (#ff6600) at 20% opacity
- Text: "y" in `COLORS.accent` (#ff6600)
- Border radius proportional to size (e.g., 6px at 32px, 24px at 256px)

## File Structure

```
apps/web/
├── app/
│   ├── icon.tsx                    # Favicon (uses shared)
│   ├── lib/
│   │   ├── constants.ts            # Existing COLORS
│   │   └── icon.ts                 # NEW: Shared generator
│   └── api/
│       └── icon/
│           └── [size]/
│               └── route.ts        # NEW: Dynamic API
├── public/
│   └── fonts/
│       └── JetBrains_Mono/
│           └── static/
│               └── JetBrainsMono-Bold.ttf
└── icon-api.yapi.yml               # NEW: Test file
```

## API Contract

### GET /api/icon/[size]

**Valid sizes**: 16, 32, 64, 128, 256

**Success Response** (200):
- Content-Type: `image/png`
- Cache-Control: `public, max-age=31536000, immutable`
- Body: PNG image data

**Error Response** (400):
```json
{
  "error": "Invalid size. Supported sizes: 16, 32, 64, 128, 256"
}
```

## Complexity Justification

| Concern | Why Needed | Simpler Alternative Rejected |
|---------|------------|------------------------------|
| Shared module | Avoids code duplication between icon.tsx and API | Inline duplication would violate DRY |
| Dynamic route | Spec requires multiple sizes via API | Static files would require manual generation |
