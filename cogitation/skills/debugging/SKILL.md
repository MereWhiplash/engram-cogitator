---
name: debugging
description: Provides systematic debugging workflow that finds root cause before fixing. Searches EC for prior issues and stores learnings. Use when encountering bugs, test failures, or unexpected behavior.
---

# Systematic Debugging

Find the root cause before attempting fixes.

**Announce:** "I'm using the debugging skill to investigate this issue."

## The Rule

```
NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST
```

## The Phases

```
Investigate → Analyze → Hypothesize → Fix
```

## Phase 1: Investigate

Before any fix attempt:

### 1. EC Search First

```
ec_search:
  query: [error message or symptom]

ec_search:
  query: [affected component]
  type: learning
```

Check if we've seen this before. Prior solutions save time.

### 2. Read Error Messages

- Don't skip past errors
- Read stack traces completely
- Note line numbers, file paths

### 3. Reproduce

- Can you trigger it reliably?
- What are the exact steps?
- Does it happen every time?

### 4. Check Recent Changes

```bash
git log --oneline -10
git diff HEAD~5
```

What changed that could cause this?

### 5. Trace Data Flow

Where does the bad value come from?
- Start at the error
- Trace backwards through the call stack
- Find the source, not the symptom

## Phase 2: Pattern Analysis

### Find Working Examples

Look for similar working code:
- Same pattern elsewhere in codebase
- Reference implementations
- Documentation examples

Search EC for established patterns:

```
ec_search:
  query: [pattern name]
  type: pattern
```

### Compare

What's different between working and broken?

## Phase 3: Hypothesis

State clearly: "I think X is the root cause because Y"

Test with the smallest possible change. One variable at a time.

**If hypothesis wrong:** Form new hypothesis. Don't pile fixes.

## Phase 4: Fix @tdd

### 1. Write Failing Test

Reproduce the bug in a test first (RED).

### 2. Implement Fix

Address the root cause, not the symptom (GREEN).

### 3. Verify @verifying

Load project config for commands:

```
ec_search:
  query: project config
  type: config
```

Run verification:
- Bug test passes
- Other tests still pass
- Original symptom gone

### 4. Store Learning

If root cause was non-obvious:

```
ec_add:
  type: learning
  area: [component]
  content: [What was wrong, why it happened, how to recognize it]
  rationale: Discovered while debugging [symptom]
```

Good learnings to store:
- Error messages that are misleading
- Subtle configuration issues
- Race conditions or timing bugs
- Framework quirks

## Red Flags

Stop and return to Phase 1 if:
- "Quick fix, investigate later"
- "Just try changing X"
- "I don't fully understand but..."
- Multiple failed fix attempts

## After 3 Failed Fixes

Question the architecture:
- Is this pattern fundamentally sound?
- Should we refactor instead of patching?

Search EC for architectural decisions:

```
ec_search:
  query: [component] architecture
  type: decision
```

Discuss before attempting fix #4.
