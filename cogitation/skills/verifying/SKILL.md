---
name: verifying
description: Ensures verification commands are run before claiming success. Checks EC for project-specific gotchas. Use before asserting tests pass, builds succeed, or bugs are fixed.
---

# Verification

Don't claim something works until you've seen the evidence.

**Announce:** "I'm using the verifying skill to confirm this works."

## The Rule

```
NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION
```

## Step 0: Load Project Config

Before running commands, get the configured commands from EC:

```
ec_search:
  query: project config
  type: config
  area: project
```

Extract: `test_command`, `lint_command`, `build_command`

**If no config:** Use sensible defaults but warn:
> "No project config found. Using default commands. Consider running @init."

## Step 1: Check for Known Gotchas

Search EC for verification issues in this area:

```
ec_search:
  query: verification gotcha
  type: learning
```

Apply any known workarounds or warnings.

## The Gate

Before claiming any success:

1. **Identify** - What command proves this claim?
2. **Run** - Execute it fresh, completely
3. **Read** - Full output, check exit code
4. **Verify** - Does output confirm the claim?
5. **Then claim** - Only now state the result

## Command Reference

Use the project-configured commands:

| Claim | Command | Fallback |
|-------|---------|----------|
| "Tests pass" | `{test_command}` | `npm test` |
| "Build succeeds" | `{build_command}` | `npm run build` |
| "Types check" | `{lint_command}` | `npx tsc --noEmit` |
| "Lints clean" | `{lint_command}` | `npm run lint` |

## Examples

| Claim | Requires | Not Sufficient |
|-------|----------|----------------|
| "Tests pass" | Test output showing 0 failures | Previous run, "should pass" |
| "Build succeeds" | Build with exit 0 | Linter passing |
| "Bug fixed" | Test for the bug passes | "I changed the code" |
| "Type-safe" | Type-check with no errors | "Looks right" |

## Forbidden Phrases

Without fresh evidence, never say:
- "Should work now"
- "That should fix it"
- "Tests probably pass"
- "I'm confident this works"

## The Pattern

```
NO: "Fixed the bug, tests should pass now"

YES: [Run tests]
     "Tests pass: 47/47. Bug fix verified."
```

## Store Gotchas

If verification reveals a non-obvious issue:

```
ec_add:
  type: learning
  area: verification
  content: [What the gotcha was]
  rationale: Discovered during verification
```

Examples worth storing:
- Tests pass locally but CI requires environment variable
- Build succeeds but output differs on ARM vs x86
- Lint passes but pre-commit hook has stricter rules

## Why This Matters

- "I don't believe you" = trust broken
- Undefined functions shipped = crashes
- Time wasted on false completion = rework

Evidence first. Claims second.
