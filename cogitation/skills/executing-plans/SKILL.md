---
name: executing-plans
description: Use when you have a written plan in docs/plans/ ready to implement — runs every task continuously via TDD red-green-refactor, then reviews
---

# Executing Plans

Execute a plan continuously in TDD, then review.

**Announce:** "I'm using the executing-plans skill to implement this plan."

## The Rule

```
Every task follows @tdd: RED → GREEN → REFACTOR
No production code without a failing test first.
```

## Prerequisites

- Plan exists in `docs/plans/YYYY-MM-DD-<topic>.md`
- On a feature branch

## The Flow

```
Verify Branch → EC Search → Load Plan → Execute Tasks → Review → Finish
```

## Step 1: Verify Branch

```bash
git branch --show-current
```

Must be on a feature branch, not main.

## Step 2: Load Context

Get project config and relevant context:

```
ec_search:
  query: project config
  type: config

ec_search:
  query: [feature area]
  type: pattern

ec_search:
  query: [feature area]
  type: learning
```

Note any gotchas or patterns that apply to this implementation.

## Step 3: Load and Review Plan

1. Read the plan file
2. **Choose ONE progress tracking approach** (don't mix):
   - **Tasks (preferred):** If TaskCreate/TaskUpdate tools are available, use Tasks. Creates persistent, shareable progress.
   - **TodoWrite (fallback):** If Tasks aren't available, use TodoWrite for the session.
3. If concerns about the plan, raise them before starting

**Important:** Pick one approach and stick with it for the entire execution. Don't mix Tasks and TodoWrite.

### Option A: Using Tasks (Preferred)

If TaskCreate/TaskUpdate/TaskList tools are available, create one task per plan task:
```
TaskCreate: "Task 1: [first task summary]"
TaskCreate: "Task 2: [second task summary]" → addBlockedBy: [task 1 id]
... one per plan task, chained in order
```

Benefits:
- Progress survives context switches and session restarts
- Subagents can share the same task list with `CLAUDE_CODE_TASK_LIST_ID`
- Dependencies prevent out-of-order execution

### Option B: Using TodoWrite (Fallback)

If Tasks aren't available, create a TodoWrite with all plan tasks and update as you go.

## Step 4: Execute the Plan

Execute **continuously** — work through every task in order, stopping only for a real blocker (see "When to Stop"). Do **not** pause for feedback between tasks; that round-trip adds latency without improving quality.

**Decide once, up front, how to run the whole plan:**
```json
{
  "questions": [{
    "question": "How should I execute this plan?",
    "header": "Execution",
    "options": [
      { "label": "Main thread", "description": "Execute here with full visibility (default)" },
      { "label": "Subagent-driven", "description": "Dispatch a fresh subagent to run the plan; returns a summary" }
    ],
    "multiSelect": false
  }]
}
```

### Main Thread Execution

For each task, follow **@tdd**:
1. Mark the task `in_progress` (TaskUpdate if using Tasks, otherwise TodoWrite)
2. **RED:** Write a failing test for the behavior this task introduces
3. **GREEN:** Write minimal production code to make it pass
4. **REFACTOR:** Clean up while keeping tests green
5. Run full verifications as specified (`@verifying`)
6. Commit after the task passes
7. Mark `completed`, then move straight to the next task

### Subagent-Driven Execution

Dispatch with this prompt:
```markdown
Execute this plan end to end:

[Paste the plan tasks]

Requirements:
- EVERY task follows @tdd: write failing test FIRST, then minimal code to pass, then refactor
- No production code without a failing test — if you can't test it, stop and report
- Use @verifying before claiming any step complete
- Commit after each task
- Execute continuously; stop and report ONLY on a blocker

EC Context:
- Test command: {test_command}
- [Relevant patterns/learnings from EC]

Return:
- Summary of what was implemented
- Tests written and their status
- Files created/modified
- Any blockers encountered
```

**Tip:** If using Tasks, the subagent can share the same task list by setting `CLAUDE_CODE_TASK_LIST_ID` — updates broadcast across sessions.

When all tasks are done and verified:
> "Implementation complete. [Brief summary]."

## Step 5: Review (the review ladder)

Pick the lightest tier that fits the change. Default to **Tier 0**.

- **Tier 0 — Inline self-review (default).** No dispatch. Re-read the full diff yourself against this checklist:
  - Spec coverage — every plan task actually implemented, nothing extra (YAGNI)
  - Tests assert real behavior, not mocks; edge cases covered
  - No placeholders / TODO / dead code left behind
  - Internal consistency — names, types, error handling match the codebase
  - Verifications pass (`@verifying`)

  Catches most issues in ~30s. For small/medium changes this is enough.

- **Tier 1 — Subagent review (opt-in).** For larger or higher-risk diffs, use `@requesting-review` to dispatch an independent reviewer.

- **Tier 2 — External / adversarial review (opt-in).** When you want a second model to pressure-test the work, and `codex` is enabled in `.cogitation/config.json`:
  - **Code / implementation:** prefer the official `codex-plugin-cc` — `/codex:adversarial-review --background` (non-blocking; retrieve with `/codex:result`).
  - **Design / plan docs:** use `@codex-review` (pre-execution artifact review).

Process feedback from any tier with `@receiving-review`. Address Critical and Important issues before finishing. Minor/style notes don't block.

## Step 6: Store Patterns

If the plan noted patterns to store:

```
ec_add:
  type: pattern
  area: [component]
  content: [Pattern description]
  rationale: Established during [feature] implementation
```

## Step 7: Finish

When all tasks complete and review passes:

> "Implementation complete. Ready to finish the branch?"

If yes → **Use @finishing-branch**

## When to Stop

**Stop and ask when:**
- Blocker mid-task (missing dependency, test fails)
- Plan has gaps
- Instruction is unclear
- Verification fails repeatedly

Don't guess - ask for clarification.
