---
name: requesting-review
description: Prepares and requests code review with proper context. Searches EC for reviewer preferences and past feedback patterns. Use before finishing a branch or when ready for feedback on implementation.
---

# Requesting Code Review

Prepare changes for review with clear context.

**Announce:** "I'm using the requesting-review skill to prepare this for review."

## The Flow

```
Verify Ready → Load Context → Prepare → Dispatch Reviewer → Handle Response
```

## Step 1: Verify Ready @verifying

Load project config:

```
ec_search:
  query: project config
  type: config
```

Before requesting review:

```bash
{test_command}
{lint_command}
{build_command}
```

**If failures:** Stop. Fix first using `@debugging`.

## Step 2: Load Review Context

Search EC for relevant review history:

```
ec_search:
  query: code review feedback
  type: learning

ec_search:
  query: review pattern
  type: pattern
```

Note any prior feedback patterns to proactively address.

## Step 3: Prepare Context

Gather what the reviewer needs:

```bash
git log --oneline main..HEAD
git diff --stat main..HEAD
```

Identify:
- **Scope**: What changed and why
- **Key decisions**: Architectural choices made
- **Risk areas**: Parts that need careful review
- **Testing**: How it was verified
- **EC context**: Prior decisions/patterns that informed this work

## Step 4: Dispatch Reviewer

Use the `code-reviewer` agent:

```markdown
Task: Code review for [feature name]
Agent: code-reviewer

## Context
[Brief description of what was implemented]

## Changes
[Output from git diff --stat]

## Key Decisions
- [Decision 1 and rationale]
- [Decision 2 and rationale]

## EC Context Consulted
- [Relevant decisions/patterns from EC]

## Areas Needing Attention
- [Specific file or pattern to scrutinize]

## Plan Reference
docs/plans/YYYY-MM-DD-<topic>.md
```

## Step 5: Handle Response

When reviewer returns:

**If issues found:**
1. Categorize: Critical / Important / Minor
2. Address Critical and Important before proceeding
3. Use `@receiving-review` to process feedback

**If approved:**
> "Code review passed. Ready to finish the branch?"

If yes → **Use @finishing-branch**

## Store Review Learnings

If review reveals a pattern worth remembering:

```
ec_add:
  type: learning
  area: code-review
  content: [What reviewers commonly catch or request]
  rationale: Recurring review feedback
```

Examples worth storing:
- Consistent feedback about error handling
- Patterns reviewers expect in this codebase
- Common oversights to check before requesting review

## What Makes a Good Review Request

| Do | Don't |
|----|-------|
| Provide context on decisions | Dump code without explanation |
| Highlight risk areas | Assume reviewer knows everything |
| Reference the plan | Skip verification before requesting |
| Keep scope focused | Request review of WIP code |
| Note EC context consulted | Ignore prior decisions |

## Integration Points

- Called by: `@executing-plans` (Step 5)
- Calls: `code-reviewer` agent, `@receiving-review`, `@finishing-branch`
