# Copilot Code Review Instructions

When performing a code review, apply the following checks:

## Team Values
- Find ways to make code more information dense or DRY, without code golfing.
- We are a team that values simplicity and clarity, but also generics to increase information density.
- Look for opportunities to simplify logic and reduce complexity
- You can't have bugs in code you don't have.

## Documentation Sync

When reviewing changes to CLI features, config schema, or user-facing functionality:

1. Check if `README.md` needs to be updated to reflect the changes
2. Check if `SKILL.md` needs to be updated with new features or usage patterns
3. Check if `apps/web/app/components/Landing.tsx` examples are still accurate
4. Verify that any new YAML config options are documented with examples

## Code Quality

- Focus on readability and avoid nested ternary operators
- Ensure error messages are actionable and user-friendly
- Check for proper error handling in async operations
- Verify that new features have corresponding example files in `examples/`
