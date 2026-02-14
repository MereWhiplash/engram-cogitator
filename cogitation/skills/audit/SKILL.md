---
name: audit
description: Review and maintain EC memories. List all memories, find duplicates, prune stale entries, and organize by type. Use when memories need cleanup, when user says "audit memories", "clean up EC", "review what's stored", or periodically for hygiene.
---

# Audit EC Memories

Review, clean, and organize stored memories.

**Announce:** "I'm using the audit skill to review EC memories."

## The Flow

```
List All → Analyze → Present → Clean → Verify
```

## Step 1: List All Memories

Pull everything by type:

```
ec_list:
  limit: 100
  type: decision

ec_list:
  limit: 100
  type: learning

ec_list:
  limit: 100
  type: pattern

ec_list:
  limit: 100
  type: config
```

Also check for invalidated entries:

```
ec_list:
  limit: 50
  include_invalid: true
```

## Step 2: Analyze

### Check for Issues

1. **Duplicates** - Same or very similar content
2. **Stale** - References outdated code/patterns
3. **Vague** - Content too generic to be useful
4. **Miscategorized** - Wrong type for the content

### Group by Area

Organize memories by their area tag to see coverage:
- Which areas have many memories?
- Which areas are sparse?
- Any orphaned areas (code deleted)?

## Step 3: Present Summary

```markdown
## EC Memory Audit

**Total Memories:** N (D decisions, L learnings, P patterns, C config)
**Invalidated:** N

### By Area
| Area | Decisions | Learnings | Patterns |
|------|-----------|-----------|----------|
| auth | 3 | 2 | 1 |
| api  | 1 | 4 | 2 |
| ...  | ... | ... | ... |

### Issues Found
- **Duplicates (N):** [list IDs]
- **Potentially stale (N):** [list with reason]
- **Vague entries (N):** [list IDs]
```

## Step 4: Clean Up

**Use AskUserQuestion** for each category:

```json
{
  "questions": [{
    "question": "How should I handle the N duplicate memories?",
    "header": "Duplicates",
    "options": [
      { "label": "Merge", "description": "Keep best version, invalidate others" },
      { "label": "Keep all", "description": "They might have nuance" },
      { "label": "Review each", "description": "Show me one by one" }
    ],
    "multiSelect": false
  }]
}
```

### Invalidating Memories

When removing a memory, use `ec_invalidate` with optional superseding:

```
ec_invalidate:
  id: <old_memory_id>
  superseded_by: <new_memory_id>  # Optional: link to replacement
```

### Merging Duplicates

1. Create new consolidated memory with `ec_add`
2. Invalidate old entries with `ec_invalidate`, pointing to new ID

## Step 5: Verify

After cleanup, show the delta:

```markdown
## Cleanup Complete

**Before:** N memories
**After:** M memories
**Invalidated:** X entries

### Changes
- Merged N duplicates
- Invalidated M stale entries
- Recategorized P entries
```

## Audit Checklist

| Check | Action |
|-------|--------|
| Duplicate content | Merge or invalidate |
| References deleted code | Invalidate |
| Too vague to act on | Rewrite or invalidate |
| Wrong type | Create new with correct type, invalidate old |
| Missing area tag | Update via new entry |

## When to Audit

- **Periodically** - Every few weeks on active projects
- **After major refactors** - Code changed, memories may be stale
- **Before onboarding** - Clean slate for new team members
- **When searches return noise** - Too many irrelevant results

## Red Flags

Stop and ask if:
- About to invalidate >10 memories
- Unsure if memory is still valid
- Memory references code you can't find
