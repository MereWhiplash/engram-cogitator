---
name: receiving-review
description: Handles code review feedback by verifying before implementing and pushing back if incorrect. Checks EC for context on prior decisions. Use when receiving code review comments or PR feedback.
---

# Receiving Code Review

Technical evaluation, not emotional performance.

**Announce:** "I'm using the receiving-review skill to process this feedback."

## The Pattern

```
Read → Understand → Check EC → Verify → Evaluate → Implement
```

### 1. Read

Complete feedback without reacting.

### 2. Understand

Restate the requirement in your own words. If unclear, ask.

### 3. Check EC

Before accepting or rejecting, check for relevant context:

```
ec_search:
  query: [topic of feedback]
  type: decision

ec_search:
  query: [component mentioned]
  type: pattern
```

The current implementation might be there for a reason.

### 4. Verify

Check against codebase reality:
- Is this technically correct for THIS codebase?
- Does it break existing functionality?
- Is there a reason for the current implementation?
- Does EC have context explaining why it's done this way?

### 5. Evaluate

Is the suggestion sound? Consider:
- Does reviewer have full context?
- Does it conflict with prior architectural decisions?
- Is it YAGNI? (grep for usage first)

### 6. Implement @tdd

One item at a time. Test each fix (`@verifying`).

## When to Push Back

Push back with technical reasoning if:
- Suggestion breaks existing functionality
- Reviewer lacks context (point to EC decision)
- Violates YAGNI (feature unused)
- Conflicts with architectural decisions stored in EC
- Technically incorrect for this stack

Example:
> "Checked EC - we decided to use X because [rationale from decision memory]. Still want to change?"

## Handling Unclear Feedback

If ANY item is unclear: **Stop. Ask first.**

Don't implement what you understand and ask about the rest. Items may be related.

## Response Format

**When feedback is correct:**
```
"Fixed. [Brief description]"
"Good catch - [issue]. Fixed in [location]."
[Just fix it, show in code]
```

**When pushing back:**
```
"Checked EC - this was decided in [area]: [rationale]. Keep current impl?"
"Grepped codebase - nothing calls this. Remove (YAGNI)?"
"This conflicts with [pattern from EC]. Want to change the pattern?"
```

**Never:**
```
"You're absolutely right!"
"Great point!"
"Thanks for catching that!"
```

Actions speak. Just fix it or explain why not.

## If You Pushed Back and Were Wrong

```
"You were right - checked [X] and it does [Y]. Fixing now."
```

State the correction factually. Move on.

## Store Learnings

If feedback reveals something worth remembering:

```
ec_add:
  type: learning
  area: [component]
  content: [What the feedback taught about this codebase]
  rationale: Learned during code review
```

Worth storing:
- Reviewer preferences that apply broadly
- Codebase conventions you weren't aware of
- Edge cases you missed that apply elsewhere
