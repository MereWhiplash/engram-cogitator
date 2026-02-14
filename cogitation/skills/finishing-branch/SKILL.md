---
name: finishing-branch
description: Verifies tests, reviews for EC memory storage, and presents merge options. Use when implementation is complete and ready to finish a feature branch.
---

# Finishing a Branch

Complete development work with verification, memory storage, and merge options.

**Announce:** "I'm using the finishing-branch skill to complete this work."

## The Flow

```
Verify Tests → EC Review → Present Options → Execute
```

## Step 1: Verify Tests @verifying

Load project config:

```
ec_search:
  query: project config
  type: config
```

Run all verifications:

```bash
{test_command}
{lint_command}
{build_command}
```

**If failures:** Stop. Fix using `@debugging` before proceeding.

**If passing:** Continue.

## Step 2: EC Review

Before finishing, review what's worth remembering.

Look at what changed:
```bash
git log --oneline main..HEAD
git diff --stat main..HEAD
```

**Use AskUserQuestion:**
```json
{
  "questions": [{
    "question": "Any decisions, learnings, or patterns worth storing in EC?",
    "header": "Memory",
    "options": [
      { "label": "Yes", "description": "I'll describe what to store" },
      { "label": "No", "description": "Nothing notable this time" },
      { "label": "Let me suggest", "description": "Claude proposes, I approve" }
    ],
    "multiSelect": false
  }]
}
```

If "Let me suggest" - analyze the changes and propose:
- **decisions** - Architectural choices made
- **learnings** - Gotchas discovered during implementation
- **patterns** - New conventions established

For each approved memory:

```
ec_add:
  type: decision|learning|pattern
  area: [component]
  content: [What to remember]
  rationale: [Why it matters]
```

## Step 3: Present Options

Load branch convention:

```
ec_search:
  query: project config branch
  type: config
```

```json
{
  "questions": [{
    "question": "How do you want to finish this branch?",
    "header": "Finish",
    "options": [
      { "label": "Create PR", "description": "Push and open pull request" },
      { "label": "Merge locally", "description": "Merge to main locally" },
      { "label": "Keep as-is", "description": "I'll handle it later" },
      { "label": "Discard", "description": "Delete this branch and work" }
    ],
    "multiSelect": false
  }]
}
```

## Step 4: Execute Choice

### Create PR (Recommended)

```bash
git push -u origin $(git branch --show-current)

gh pr create --title "<title>" --body "$(cat <<'EOF'
## Summary
- [What changed]

## Test Plan
- [ ] [Verification steps]

## EC Context
- [Relevant decisions/patterns consulted or created]

---
Design: docs/designs/YYYY-MM-DD-<topic>.md
Plan: docs/plans/YYYY-MM-DD-<topic>.md
EOF
)"
```

### Merge Locally

```bash
git checkout main
git pull
git merge <feature-branch>
{test_command}  # Verify on merged result
git branch -d <feature-branch>
```

### Keep As-Is

> "Branch preserved. You can return to it later."

### Discard

**Confirm first:**
> "This will delete branch `<name>` and all commits. Type 'discard' to confirm."

```bash
git checkout main
git branch -D <feature-branch>
```

## Post-Finish

After successful merge/PR:

> "Branch complete. What's next?"

Suggest based on context:
- More features → **Use @brainstorming**
- Bug to fix → **Use @debugging**
- Take a break → Session summary with EC memories stored
