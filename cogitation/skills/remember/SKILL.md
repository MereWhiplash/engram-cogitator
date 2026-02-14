---
name: remember
description: Store important project context in EC's persistent memory. Use after making architectural decisions, discovering gotchas or workarounds, identifying recurring patterns, or when user says "remember this", "store this", "save for later", "don't forget".
---

# Storing Memories

Use the 2-of-3 rule. Store if at least TWO are true:
1. **Future sessions** - Will this matter when returning to this project?
2. **Project-specific** - Is this unique to this codebase?
3. **Costly to rediscover** - Would finding this again waste time or tokens?

## Memory Types

| Type | Use for | Examples |
|------|---------|----------|
| `decision` | Architectural choices, why X over Y | "Chose Redis over Memcached because..." |
| `learning` | Codebase discoveries, gotchas, workarounds | "API rate limits are 100 req/min..." |
| `pattern` | Recurring conventions in this project | "All API responses use envelope format..." |
| `config` | Project configuration (managed by @init/@config) | Test commands, branching conventions |

## Before Storing

### 1. Search First

```
ec_search:
  query: [what you want to store]
```

Avoid duplicates. If similar exists, consider updating instead.

### 2. Evaluate Worth

**Store:**
- Architectural decisions with rationale
- Non-obvious gotchas that cost time
- Project-specific conventions
- Integration quirks with external services

**Skip:**
- Obvious things (syntax, common patterns)
- Temporary fixes or WIP
- Things easily found in docs
- Generic programming knowledge

## Storing

Use `ec_add` with required fields:

```
ec_add:
  type: decision|learning|pattern
  area: [component like "auth", "api", "database"]
  content: [1-2 sentences, specific and actionable]
  rationale: [Why this matters, optional but recommended]
```

## Good vs Bad Examples

### Decisions

```
GOOD:
  type: decision
  area: auth
  content: Using JWT with short-lived access tokens (15min) and refresh tokens (7d) for session management
  rationale: Balance between security (short access) and UX (don't require frequent re-login)

BAD:
  type: decision
  area: auth
  content: Using JWT for authentication
  rationale: It's popular
```

### Learnings

```
GOOD:
  type: learning
  area: api
  content: External payment API returns 200 with error in body for validation failures. Must check response.success field, not just HTTP status.
  rationale: Discovered during integration - caused silent failures initially

BAD:
  type: learning
  area: api
  content: API sometimes fails
```

### Patterns

```
GOOD:
  type: pattern
  area: components
  content: All form components use react-hook-form with zod validation. Schema defined in same file as component.
  rationale: Established convention for consistency

BAD:
  type: pattern
  area: components
  content: We use React
```

## When to Store

Store immediately after:
- Making an architectural decision
- Discovering a non-obvious issue
- Establishing a new convention
- Solving a tricky bug (store the root cause)
- Learning something project-specific

## Updating Memories

If a memory is outdated:

1. Create new memory with current information
2. Invalidate old one:

```
ec_invalidate:
  id: <old_memory_id>
  superseded_by: <new_memory_id>
```

## Quick Reference

The `area` field should match component/domain names used in your codebase:
- `auth`, `api`, `database`, `ui`, `config`, `testing`, `deployment`
- Use consistent naming across memories for better search results
