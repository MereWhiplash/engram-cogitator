---
name: cog-init
description: Initialize a project for cogitation workflows. Sets up EC project context, scaffolds directories, and configures project-specific commands.
---

# Initialize Cogitation

First-time setup for a project. Creates the foundation for memory-enhanced workflows.

**Announce:** "I'm using the init skill to set up cogitation for this project."

## The Flow

```
Verify EC → Check Existing → Scaffold → Configure → Store
```

## Step 1: Verify EC Connection

Test that Engram Cogitator is running:

```
ec_search: test connection
```

**If EC unavailable:** Stop. Cogitation requires EC.

> "EC connection failed. Please ensure engram-cogitator is running. See: https://github.com/MereWhiplash/engram-cogitator"

## Step 2: Check for Existing Context

Search for prior initialization:

```
ec_search: project config
ec_search: cogitation initialized
```

**If found:** Ask before reinitializing.

```json
{
  "questions": [{
    "question": "This project already has cogitation context. What would you like to do?",
    "header": "Existing",
    "options": [
      { "label": "Keep existing", "description": "Exit without changes" },
      { "label": "Reconfigure", "description": "Update settings, keep memories" },
      { "label": "Fresh start", "description": "Clear all and reinitialize" }
    ],
    "multiSelect": false
  }]
}
```

## Step 3: Scaffold Directories

Create standard directories if they don't exist:

```bash
mkdir -p docs/designs docs/plans
```

## Step 4: Detect Project Stack

Analyze the project to suggest defaults:

```bash
ls -la
```

Look for:
- `package.json` → Node/npm/pnpm/bun
- `go.mod` → Go
- `Cargo.toml` → Rust
- `pyproject.toml` / `requirements.txt` → Python
- `Makefile` → Make-based

## Step 5: Configure Commands

**Use AskUserQuestion** to confirm/customize commands:

```json
{
  "questions": [
    {
      "question": "What command runs your tests?",
      "header": "Test",
      "options": [
        { "label": "[detected]", "description": "Based on project files" },
        { "label": "pnpm test", "description": "pnpm package manager" },
        { "label": "npm test", "description": "npm package manager" },
        { "label": "go test ./...", "description": "Go testing" }
      ],
      "multiSelect": false
    },
    {
      "question": "What command checks types/lints?",
      "header": "Lint",
      "options": [
        { "label": "[detected]", "description": "Based on project files" },
        { "label": "pnpm lint", "description": "ESLint via pnpm" },
        { "label": "go vet ./...", "description": "Go vet" },
        { "label": "None", "description": "Skip linting" }
      ],
      "multiSelect": false
    },
    {
      "question": "What command builds the project?",
      "header": "Build",
      "options": [
        { "label": "[detected]", "description": "Based on project files" },
        { "label": "pnpm build", "description": "Build via pnpm" },
        { "label": "go build ./...", "description": "Go build" },
        { "label": "None", "description": "No build step" }
      ],
      "multiSelect": false
    }
  ]
}
```

## Step 6: Configure Branching

```json
{
  "questions": [{
    "question": "What branching convention do you use?",
    "header": "Branches",
    "options": [
      { "label": "feat/fix/chore", "description": "feat/<name>, fix/<name>, chore/<name>" },
      { "label": "feature/bugfix", "description": "feature/<name>, bugfix/<name>" },
      { "label": "Flat", "description": "Just descriptive names" }
    ],
    "multiSelect": false
  }]
}
```

## Step 7: Store Configuration

Store the project configuration in EC:

```
ec_add:
  type: config
  area: project
  content: |
    test_command: <user choice>
    lint_command: <user choice>
    build_command: <user choice>
    branch_convention: <user choice>
  rationale: Project configuration for cogitation workflows
```

Mark initialization complete:

```
ec_add:
  type: pattern
  area: project
  content: Cogitation initialized for this project
  rationale: Marker for init detection
```

## Step 8: Summary

Present what was set up:

> **Cogitation initialized**
>
> - Test: `<command>`
> - Lint: `<command>`
> - Build: `<command>`
> - Branches: `<convention>`
>
> Directories created: `docs/designs/`, `docs/plans/`
>
> Ready to use cogitation skills. Try `@brainstorming` to start a feature.

## Handoff

Suggest next steps based on context:

- New feature idea → **Use @brainstorming**
- Existing bug → **Use @debugging**
- Want to understand workflows → List available skills
