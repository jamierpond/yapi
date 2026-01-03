# Tasks: Update App Icon

**Branch**: `001-update-app-icon` | **Plan**: [plan.md](./plan.md)

## Phase 1: Core Implementation

- [x] **T1**: Create shared icon generator module
  - File: `apps/web/app/lib/icon.ts`
  - Export `generateIcon(size: number)` function
  - Load JetBrains Mono Bold font
  - Render "y" with gradient background using COLORS constants
  - Scale font and border-radius proportionally

- [x] **T2**: Update favicon to use shared generator
  - File: `apps/web/app/icon.tsx`
  - Import `generateIcon` from `./lib/icon`
  - Update size export to 256x256
  - Call `generateIcon(256)`

- [x] **T3**: Create dynamic API route
  - File: `apps/web/app/api/icon/[size]/route.ts`
  - Parse size from params
  - Validate against allowed sizes (16, 32, 64, 128, 256)
  - Return 400 for invalid sizes
  - Add cache headers
  - Call `generateIcon(size)`

## Phase 2: Testing & Validation

- [x] **T4**: Create yapi test file
  - File: `apps/web/icon-api.yapi.yml`
  - Test all valid sizes with output_file
  - Test invalid size returns 400

- [ ] **T5**: Manual verification
  - Run dev server
  - Test each endpoint visually
  - Verify favicon appears correctly
  - Run yapi test file

- [x] **T6**: Build verification
  - Run `pnpm build` in apps/web
  - Ensure no build errors

## Execution Order

```
T1 → T2 → T3 → T4 → T5 → T6
     (sequential, each depends on previous)
```

## Acceptance Criteria

- [ ] `/api/icon/256` returns 256x256 PNG
- [ ] `/api/icon/32` returns 32x32 PNG
- [ ] `/api/icon/999` returns 400 error
- [ ] Favicon displays at 256x256
- [ ] Icon uses JetBrains Mono Bold font
- [ ] Icon has gradient background matching landing page style
- [ ] yapi test file passes all assertions
- [ ] Build completes successfully
