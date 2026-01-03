# Feature Specification: Update App Icon

**Branch**: `001-update-app-icon` | **Created**: 2026-01-03 | **Status**: Draft

## Overview

Update the web application's favicon/icon system to support multiple sizes via a dynamic API endpoint (`/api/icon/{size}`), use the project's JetBrains Mono font, incorporate landing page-inspired gradient styling, and share the icon generation logic between the favicon and API routes.

## User Stories

### US1 - Improved Brand Recognition (P1)

As a user visiting the website, I want to see a crisp, high-resolution favicon in my browser tab and bookmarks so that I can easily identify the yapi website among my open tabs and saved bookmarks.

**Acceptance**:
- Given a user opens the yapi website, when they view the browser tab, then they see a clear, sharp icon
- Given a user bookmarks the website, when they view their bookmarks, then the icon appears without pixelation

---

### US2 - Dynamic Icon Sizes (P1)

As a developer or integration, I want to request icons at specific sizes via an API so that I can use appropriately sized icons for different contexts (social sharing, app icons, thumbnails).

**API Usage**:
```
GET /api/icon/16
GET /api/icon/32
GET /api/icon/128
GET /api/icon/256
```

**Acceptance**:
- Given a request to `/api/icon/32`, when the response is received, then it returns a 32x32 PNG image
- Given a request to `/api/icon/256`, when the response is received, then it returns a 256x256 PNG image
- Given a request to `/api/icon/999`, when the response is received, then it returns an appropriate error or redirects to a default size

---

### US3 - Consistent Visual Identity (P2)

As a user, I want the favicon to match the visual style of the landing page so that the brand feels cohesive across all touchpoints.

**Acceptance**:
- Given a user views the favicon alongside the landing page, when comparing visual elements, then the icon uses the same font (JetBrains Mono) and similar color scheme/gradients as the landing page

## Requirements

### Functional

- **FR-001**: The system MUST provide an API endpoint at `/api/icon/{size}` that returns PNG icons at the requested pixel size
- **FR-002**: The icon.tsx favicon MUST use the same shared icon generation logic as the API
- **FR-003**: The icon MUST use the JetBrains Mono Bold font for the "y" character
- **FR-004**: The icon MUST incorporate a gradient background inspired by the landing page design
- **FR-005**: The API MUST support common icon sizes (16, 32, 64, 128, 256)
- **FR-006**: The API MUST validate the size parameter and return an error for invalid/unsupported sizes
- **FR-007**: The default favicon size MUST be 256x256
- **FR-008**: A yapi test file MUST be provided to validate and download icons from the API

### Testing with yapi

A test file (`icon-api.yapi.yml`) should be created to validate and download icons:

```yaml
yapi: v1
chain:
  - name: icon_256
    url: ${url}/api/icon/256
    method: GET
    output_file: ./icons/icon-256.png
    expect:
      status: 200
      headers:
        content-type: image/png

  - name: icon_32
    url: ${url}/api/icon/32
    method: GET
    output_file: ./icons/icon-32.png
    expect:
      status: 200

  - name: icon_16
    url: ${url}/api/icon/16
    method: GET
    output_file: ./icons/icon-16.png
    expect:
      status: 200

  - name: icon_invalid
    url: ${url}/api/icon/999
    method: GET
    expect:
      status: 400
```

Run with: `yapi run icon-api.yapi.yml -e local` to download icons for visual inspection.

### Visual Design

The icon should incorporate visual elements from the landing page design:
- Primary accent color gradient (orange/amber tones)
- Dark background consistent with the site theme
- The "y" character prominently displayed
- Subtle visual depth or glow effect to match the landing page aesthetic

### Supported Sizes

| Size | Use Case |
|------|----------|
| 16   | Browser tab favicon |
| 32   | Bookmark icons, standard favicon |
| 64   | Small thumbnails |
| 128  | Medium displays |
| 256  | High-resolution, social sharing |

## Edge Cases

- What happens when an unsupported size is requested (e.g., 999)?
- How does the icon scale the "y" character proportionally for very small sizes?
- What happens if the font file fails to load?
- Should there be caching headers on the API response?

## Success Criteria

- [ ] API endpoint `/api/icon/{size}` returns correctly sized PNG images
- [ ] icon.tsx and API share the same icon generation logic
- [ ] Icon uses JetBrains Mono Bold font consistently
- [ ] Icon background incorporates gradient styling from landing page
- [ ] Icon remains recognizable when scaled down to 16x16
- [ ] Invalid size requests return appropriate error responses
- [ ] yapi test file passes all assertions and downloads icons successfully
- [ ] Build process completes successfully

## Assumptions

- The existing JetBrains Mono Bold font file at `public/fonts/JetBrains_Mono/static/JetBrainsMono-Bold.ttf` will be used
- The COLORS constants from the existing codebase define the accent and background colors
- Caching can use standard browser caching headers (no CDN configuration needed)
- The gradient styling should be inspired by the landing page's glowing orb and accent color effects

## Out of Scope

- SVG icon format support
- Adding animation to the icon
- Creating separate icons for different themes (dark/light mode)
- Apple Touch Icon or other platform-specific icon metadata files
