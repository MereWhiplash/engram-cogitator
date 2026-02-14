---
name: config
description: View or update cogitation project configuration. Modify test/build/lint commands, branching conventions, or other project settings. Use when user says "change config", "update settings", "set test command", or wants to modify project setup.
---

# Configure Cogitation

View or update project configuration stored in EC.

**Announce:** "I'm using the config skill to manage project settings."

## The Flow

```
Load Current → Present → Modify → Save → Verify
```

## Step 1: Load Current Configuration

```
ec_search:
  query: project config
  type: config
  area: project
```

**If no config found:**

> "No configuration found. Want to initialize cogitation for this project?"

If yes → **Use @init**

## Step 2: Present Current Settings

```markdown
## Current Configuration

| Setting | Value |
|---------|-------|
| Test command | `pnpm test` |
| Lint command | `pnpm lint` |
| Build command | `pnpm build` |
| Branch convention | `feat/fix/chore` |

What would you like to change?
```

## Step 3: Modify Settings

**Use AskUserQuestion:**

```json
{
  "questions": [{
    "question": "Which setting do you want to modify?",
    "header": "Setting",
    "options": [
      { "label": "Test command", "description": "Command to run tests" },
      { "label": "Lint command", "description": "Command to check types/lint" },
      { "label": "Build command", "description": "Command to build project" },
      { "label": "Branch convention", "description": "How branches are named" }
    ],
    "multiSelect": true
  }]
}
```

For each selected setting, ask for the new value:

```json
{
  "questions": [{
    "question": "What should the test command be?",
    "header": "Test",
    "options": [
      { "label": "pnpm test", "description": "pnpm package manager" },
      { "label": "npm test", "description": "npm package manager" },
      { "label": "go test ./...", "description": "Go testing" },
      { "label": "cargo test", "description": "Rust testing" }
    ],
    "multiSelect": false
  }]
}
```

## Step 4: Save Updated Configuration

1. **Invalidate old config:**
```
ec_invalidate:
  id: <old_config_id>
```

2. **Store new config:**
```
ec_add:
  type: config
  area: project
  content: |
    test_command: <new value>
    lint_command: <new value>
    build_command: <new value>
    branch_convention: <new value>
  rationale: Updated project configuration
```

## Step 5: Verify

> **Configuration Updated**
>
> | Setting | Old | New |
> |---------|-----|-----|
> | Test | `pnpm test` | `npm test` |
>
> Changes take effect immediately for all cogitation skills.

## Quick Commands

For fast updates, users can specify directly:

- "Set test command to `go test ./...`"
- "Change lint to `npm run lint`"
- "Use feat/bugfix branching"

Parse the intent and update accordingly without full Q&A flow.

## Configuration Reference

### Test Commands
| Stack | Common Commands |
|-------|-----------------|
| Node | `npm test`, `pnpm test`, `yarn test`, `bun test` |
| Go | `go test ./...`, `go test -v ./...` |
| Rust | `cargo test` |
| Python | `pytest`, `python -m pytest` |

### Lint Commands
| Stack | Common Commands |
|-------|-----------------|
| Node | `npm run lint`, `pnpm lint`, `eslint .` |
| Go | `go vet ./...`, `golangci-lint run` |
| Rust | `cargo clippy` |
| Python | `ruff check .`, `flake8` |

### Build Commands
| Stack | Common Commands |
|-------|-----------------|
| Node | `npm run build`, `pnpm build`, `tsc` |
| Go | `go build ./...`, `go build ./cmd/...` |
| Rust | `cargo build` |
| Python | Usually none needed |

### Branch Conventions
| Style | Example |
|-------|---------|
| feat/fix/chore | `feat/add-auth`, `fix/login-bug`, `chore/update-deps` |
| feature/bugfix | `feature/user-auth`, `bugfix/session-timeout` |
| Flat | `add-user-authentication`, `fix-login-issue` |
