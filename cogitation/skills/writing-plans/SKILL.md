---
name: writing-plans
description: Creates detailed implementation plans with TDD steps from design documents. Searches EC for relevant patterns. Use after brainstorming to write an implementation plan.
---

# Writing Plans

Create implementation plans that an engineer with zero context can follow.

**Announce:** "I'm using the writing-plans skill to create the implementation plan."

## Prerequisites

- Design document exists in `docs/designs/YYYY-MM-DD-<topic>.md`
- On a feature branch

## Step 0: Load Context

Get project config and relevant patterns:

```
ec_search:
  query: project config
  type: config

ec_search:
  query: [feature area] pattern
  type: pattern

ec_search:
  query: [feature area] implementation
  type: learning
```

Note the configured `test_command` for TDD steps.

## The Plan Structure

Plans go in `docs/plans/YYYY-MM-DD-<topic>.md` and cross-reference the design.

### Header Template

```markdown
# [Feature Name] Implementation Plan

**Design:** [docs/designs/YYYY-MM-DD-<topic>.md](../designs/YYYY-MM-DD-<topic>.md)

**Goal:** [One sentence]

**Architecture:** [2-3 sentences about approach]

**EC Context:** [List any relevant memories consulted]

---
```

### Task Structure

Each task is bite-sized (2-5 minutes) and follows `@tdd`:

```markdown
### Task N: [Component Name] @tdd

**Files:**
- Create: `exact/path/to/file.ts`
- Modify: `exact/path/to/existing.ts`
- Test: `tests/path/to/test.ts`

**Step 1: Write failing test** (RED)
\`\`\`typescript
test('specific behavior', () => {
  // exact test code
});
\`\`\`

**Step 2: Run test, verify it fails**
\`\`\`bash
{test_command} path/to/test
\`\`\`
Expected: FAIL - "function not defined"

**Step 3: Implement minimal code** (GREEN)
\`\`\`typescript
// exact implementation code
\`\`\`

**Step 4: Run test, verify it passes** @verifying
\`\`\`bash
{test_command} path/to/test
\`\`\`
Expected: PASS

**Step 5: Commit**
\`\`\`bash
git add . && git commit -m "feat: add specific feature"
\`\`\`
```

## Principles

- **Exact file paths** - No ambiguity
- **Complete code** - Not "add validation here"
- **TDD every task** - Test → Fail → Implement → Pass → Commit
- **Frequent commits** - One per task
- **DRY/YAGNI** - Don't over-engineer

## EC Integration

Reference patterns found during search:

```markdown
**EC Context:**
- Pattern: [area] - [brief description of relevant pattern]
- Learning: [area] - [gotcha to be aware of]
```

If the plan establishes a new pattern, note it for storage after implementation:

```markdown
**Patterns to Store:**
- [Description of new pattern established]
```

## Handoff

After saving the plan:

> "Plan saved to `docs/plans/YYYY-MM-DD-<topic>.md`. Ready to execute?"

If yes → **Use @executing-plans**
