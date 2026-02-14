---
name: parallel-agents
description: Dispatches multiple agents for independent tasks that can run concurrently. Searches EC for prior parallelization strategies. Use when facing 2+ independent problems like separate test failures or unrelated bugs.
---

# Parallel Agents

Dispatch multiple agents for independent problems.

**Announce:** "I'm using the parallel-agents skill to handle these independent tasks."

## When to Use

**Use when:**
- 3+ test files failing with different root causes
- Multiple independent bugs/features
- No shared state between the work

**Don't use when:**
- Failures might be related (fix one = fix others)
- Agents would edit the same files
- Need to understand full context first

## Step 0: Load Context

Get project config and check for prior strategies:

```
ec_search:
  query: project config
  type: config

ec_search:
  query: parallel agent strategy
  type: learning
```

## The Pattern

### 1. Identify Independent Domains

Group problems by what's broken:
- File A tests: One subsystem
- File B tests: Different subsystem
- File C tests: Another subsystem

If fixing A might fix B, they're not independent.

### 2. Check for Shared Dependencies

Search EC for known coupling:

```
ec_search:
  query: [component A] depends
  type: learning

ec_search:
  query: [component B] coupling
  type: pattern
```

If EC indicates these components are coupled, don't parallelize.

### 3. Create Focused Tasks

Each agent gets:
- **Specific scope** - One file or subsystem
- **Clear goal** - Make these tests pass
- **Constraints** - Don't change other code
- **EC context** - Relevant patterns/learnings for that area
- **Expected output** - Summary of findings and fixes

### 4. Set Up Progress Tracking

**Choose ONE approach** (don't mix Tasks and TodoWrite):

**Option A: Tasks (Preferred)**
If TaskCreate/TaskUpdate tools are available:
```
TaskCreate: "Fix file-a.test failures"
TaskCreate: "Fix file-b.test failures"
TaskCreate: "Fix file-c.test failures"
```
No dependencies needed since these are independent. Agents can share the task list via `CLAUDE_CODE_TASK_LIST_ID` - updates broadcast in real-time.

**Option B: TodoWrite (Fallback)**
If Tasks aren't available, use TodoWrite to track each workstream.

### 5. Dispatch in Parallel

```
Task("Fix file-a.test failures")
Task("Fix file-b.test failures")
Task("Fix file-c.test failures")
```

All three run concurrently.

### 6. Review and Integrate

When agents return:
1. Read each summary
2. Check for conflicts (same files edited?)
3. Run full test suite
4. Integrate changes
5. Mark Tasks as completed (if using Tasks)

## Agent Prompt Template

```markdown
Fix the failing tests in [file]:

Failures:
1. [test name] - [error summary]
2. [test name] - [error summary]

EC Context:
- Test command: {test_command}
- [Relevant patterns for this component]
- [Known gotchas for this area]

Your task:
1. Read the test file
2. Identify root cause (@debugging)
3. Fix the issue (@tdd)
4. Verify tests pass (@verifying)

Constraints:
- Only modify files in [scope]
- Don't change [shared files]

Return: Summary of what you found and fixed.
```

## Red Flags

**Never:**
- Dispatch agents for related failures
- Let agents edit the same files
- Skip the integration verification
- Trust agent success without checking

## After Agents Return @verifying

Run full suite to verify no conflicts:

```bash
{test_command}
```

If conflicts exist, resolve manually.

## Store Learnings

If parallelization revealed something useful:

```
ec_add:
  type: learning
  area: parallel-agents
  content: [What worked or didn't work about this parallelization]
  rationale: Discovered during parallel task execution
```

Worth storing:
- Components that seemed independent but weren't
- Effective task boundaries for this codebase
- Agent prompt patterns that worked well
