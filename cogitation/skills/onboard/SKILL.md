---
name: onboard
description: Use at the start of a session on an existing project, when joining a project, or the user says "catch me up", "what do I need to know", or "onboard me"
---

# Onboard to Project

Pull institutional knowledge from EC to get context on an existing project.

**Announce:** "I'm using the onboard skill to pull project context from EC."

## The Flow

```
Verify EC → Load Config → Pull Memories → Summarize → Ready
```

## Step 1: Verify EC Connection

```
ec_search: test connection
```

**If EC unavailable:** Stop. Cannot onboard without EC.

## Step 2: Load Project Configuration

```
ec_search: project config
```

Extract:
- Test command
- Lint command
- Build command
- Branch convention

**If no config found:**

> "This project hasn't been initialized for cogitation. Want me to set it up?"

If yes → **Use @cog-init**

## Step 3: Pull All Memories

Search each memory type:

```
ec_search: type:decision
ec_search: type:learning
ec_search: type:pattern
```

## Step 3.5: Graphify Structural Recon (STRICT when enabled)

Read `.cogitation/config.json`. **If `graphify.enabled` is `true`, structural recon is REQUIRED — do not skip it.**

1. Ensure a graph exists (the AST build needs no model, no API key):
   ```bash
   test -f graphify-out/graph.json || graphify update .
   ```
2. Orient on structure:
   ```bash
   graphify query "What are the main components and how do they connect?" --budget 800
   ```
3. For specific symbols later, use `graphify query "..."` / `graphify explain "<symbol>"`.

The two sources are complementary: **EC** = decisions/gotchas/patterns; **graphify** = code structure/call-graph. Fold both into the briefing. If `graphify.enabled` is absent or `false`, skip this step entirely.

## Step 4: Categorize and Summarize

Organize findings by area:

### Architectural Decisions
List decisions with their rationale. Group by component/area.

### Learnings (Gotchas & Workarounds)
List discovered issues and their solutions. Highlight anything that would be costly to rediscover.

### Patterns (Conventions)
List established patterns and conventions specific to this project.

## Step 5: Present Summary

Format as a briefing:

> **Project Context Loaded**
>
> **Configuration:**
> - Test: `<command>`
> - Lint: `<command>`
> - Build: `<command>`
>
> **Key Decisions (N):**
> - [Area]: [Decision summary]
> - ...
>
> **Gotchas to Know (N):**
> - [Area]: [Learning summary]
> - ...
>
> **Conventions (N):**
> - [Pattern summary]
> - ...
>
> Ready to work. What would you like to do?

## Step 6: Offer Next Steps

```json
{
  "questions": [{
    "question": "What would you like to do?",
    "header": "Next",
    "options": [
      { "label": "Start a feature", "description": "Use @brainstorming" },
      { "label": "Fix a bug", "description": "Use @debugging" },
      { "label": "Review memories", "description": "Use @audit" },
      { "label": "Just exploring", "description": "I'll ask when ready" }
    ],
    "multiSelect": false
  }]
}
```

## When to Use

- **New session** - At start of conversation to load context
- **After break** - Returning to project after time away
- **New team member** - Onboarding someone unfamiliar with decisions
- **Context refresh** - When you need to remember what was decided

## What Makes Good Onboarding

| Do | Don't |
|----|-------|
| Summarize, don't dump | List every memory verbatim |
| Highlight gotchas prominently | Bury important warnings |
| Group by area/component | Random ordering |
| Note what's missing | Assume completeness |
