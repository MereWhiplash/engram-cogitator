---
name: executing-plans
description: Executes implementation plans in batches with review checkpoints. Loads EC context before each batch. Use when ready to implement a plan from docs/plans/.
---

# Executing Plans

Execute plans in batches with review checkpoints.

**Announce:** "I'm using the executing-plans skill to implement this plan."

## Prerequisites

- Plan exists in `docs/plans/YYYY-MM-DD-<topic>.md`
- On a feature branch

## The Flow

```
Verify Branch → EC Search → Load Plan → Execute Batches → Finish
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

If TaskCreate/TaskUpdate/TaskList tools are available:
```
TaskCreate: "Batch 1: [first 3 tasks summary]"
TaskCreate: "Batch 2: [next 3 tasks summary]" → addBlockedBy: [batch 1 id]
TaskCreate: "Batch 3: [final tasks summary]" → addBlockedBy: [batch 2 id]
```

Benefits:
- Progress survives context switches and session restarts
- Subagents can share the same task list with `CLAUDE_CODE_TASK_LIST_ID`
- Dependencies prevent out-of-order execution

### Option B: Using TodoWrite (Fallback)

If Tasks aren't available, create a TodoWrite with all batches and update as you go.

## Step 4: Execute in Batches

**Default batch size: 3 tasks**

**Before each batch, ask:**
```json
{
  "questions": [{
    "question": "How should I execute Batch N (tasks X-Y)?",
    "header": "Execution",
    "options": [
      { "label": "Main thread", "description": "Execute here with full visibility" },
      { "label": "Subagent", "description": "Fresh context, returns summary" }
    ],
    "multiSelect": false
  }]
}
```

### Main Thread Execution

For each task:
1. Mark batch as `in_progress` (TaskUpdate if using Tasks, otherwise TodoWrite)
2. Follow each step exactly as written (tasks marked `@tdd` use that skill)
3. Run verifications as specified (`@verifying`)
4. Mark as `completed` when batch passes verification

### Subagent Execution

Dispatch with this prompt:
```markdown
Execute tasks X-Y from this plan:

[Paste relevant task sections from plan]

Requirements:
- Follow @tdd for each task (RED → GREEN → REFACTOR)
- Use @verifying before claiming any step complete
- Commit after each task
- Stop and report if any verification fails

EC Context:
- Test command: {test_command}
- [Relevant patterns/learnings from EC]

Return:
- Summary of what was implemented
- Files created/modified
- Any issues or blockers encountered
```

After subagent returns:
1. Review the summary
2. Update progress tracking (TaskUpdate if using Tasks, otherwise TodoWrite)
3. If issues reported, switch to main thread for fixes

**Tip:** If using Tasks, subagents can share the same task list by setting `CLAUDE_CODE_TASK_LIST_ID` - updates broadcast across sessions.

**After each batch:**
> "Completed tasks N-M. [Brief summary]. Ready for feedback."

Wait for feedback before continuing.

## Step 5: Request Code Review @requesting-review

After all tasks complete, use `@requesting-review` to:
1. Verify tests pass
2. Prepare context for reviewer
3. Dispatch `code-reviewer` agent
4. Handle feedback with `@receiving-review`

Address any Critical or Important issues before proceeding.

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
- Blocker mid-batch (missing dependency, test fails)
- Plan has gaps
- Instruction is unclear
- Verification fails repeatedly

Don't guess - ask for clarification.
