---
name: ec-remember
description: Store important project context in Engram Cogitator's persistent memory. Use after making architectural decisions, discovering codebase gotchas or workarounds, identifying recurring patterns, or when user states preferences. Triggers: "remember this", "store this", "save for later", "don't forget", architectural decision, learned something, discovered a gotcha.
---

# Storing Memories with Engram Cogitator

Use the 2-of-3 rule. Store if at least TWO are true:
1. **Future sessions** - Will this matter when returning to this project?
2. **Project-specific** - Is this unique to this codebase?
3. **Costly to rediscover** - Would finding this again waste time or tokens?

## Memory Types

| Type | Use for |
|------|---------|
| `decision` | Architectural choices, why X over Y |
| `learning` | Codebase discoveries, gotchas, workarounds |
| `pattern` | Recurring conventions in this project |

## Storing

Use `ec_add` with: type, area (component like "auth", "api"), content (1-2 sentences), optional rationale.

## Before Storing

1. Search first with `ec_search` - avoid duplicates
2. Skip obvious things (syntax, common patterns)
3. Skip temporary fixes or WIP

## Good vs Bad

- ✓ "API rate limits are 100 req/min per user, discovered during load testing"
- ✗ "Fixed a typo in the README"
