# Development Workflow for Claude

## Critical Rule: Always Run Make

**BEFORE committing ANY changes, you MUST run:**

```bash
make
```

This runs:
- `fmt` - Format code
- `vet` - Check for issues
- `lint` - Run linter
- `test` - Run all tests
- `build` - Build the binary

## Development Steps

1. Make your changes
2. **Run `make`** ← NEVER SKIP THIS
3. Fix any errors from make
4. **Run `make` again** to verify fixes
5. Commit and push

## Quick Reference

```bash
# The one command you need
make

# If you want CI checks (stricter)
make ci
```

## That's It

Don't overcomplicate it. Just run `make` before committing.
