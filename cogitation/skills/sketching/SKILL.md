---
name: sketching
description: Use for MEDIUM-sized work — new but contained behavior (one component, no architecture/data/API change) that deserves rigor without the design-doc + plan-doc ceremony of brainstorming
---

# Sketching

The MEDIUM path: all of the discipline of the full flow, a fraction of the
ceremony. One half-page combo doc replaces the design doc + plan pair; the
execution rigor (TDD, verification, inline review) is identical.

**Announce:** "I'm using the sketching skill — MEDIUM-sized work, combo doc
instead of full design + plan."

## When This Fits

New behavior that is **contained**: one component, reversible-ish, no
architecture/data-model/API-surface change. Too big for straight `@tdd`
(there are real decisions to record), too small for `brainstorming` →
`writing-plans` (two ceremony docs would outweigh the change).

## The Flow

```
Quick Recon → Combo Doc → One Confirmation → Execute (TDD) → Persist
```

## Step 1: Quick Recon

Lightweight — not the full brainstorming recon:

```
ec_search:
  query: [feature area]
```

If graphify is enabled in `.cogitation/config.json`, one structural query on
the touched area. Skip anything that doesn't sharpen the sketch.

## Step 2: The Combo Doc

Write `docs/sketches/YYYY-MM-DD-<topic>.md`, capped at **half-page**:

```markdown
# <Topic>

**What/Why:** 2–3 sentences — the change and what it buys.
**Approach:** the chosen approach, and the alternative rejected (one line on why).
**Risks/Edges:** failure modes, empty/error states worth guarding.

## Tasks
- [ ] 1. <first task — becomes a failing test>
- [ ] 2. <...2–5 items total, dependency order>
```

The task checklist doubles as the plan — tick items off as they complete, so a
dead session can resume from the doc.

## Step 3: One Confirmation Round

Show the combo doc; get one confirmation from the user. This is a sanity
check, not an interview — no question loop. Fold in corrections and go.

## Step 4: Execute Continuously

Every task follows **@tdd** (red → green → refactor), with **@verifying**
before claiming completion and Tier 0 inline self-review at the end — the same
discipline as `executing-plans`, none of its ceremony. Commit per task; tick
the checklist as you go.

## Escalate When It Grows

Scope growth is a re-size, not a footnote. If mid-sketch you hit a one-way
door, a new dependency, a second component, or the checklist wants a 6th task:
**STOP, announce the re-size** ("Sizing: full — <reason>"), and escalate to
`brainstorming`, handing over the combo doc as context. Never quietly keep
sketching a FULL-sized change.

## Step 5: Persist

Durable decisions go to EC:

```
ec_add:
  type: decision
  area: [component]
  content: [What was decided and why]
  rationale: [Trade-off — including the rejected alternative]
```

The combo doc is the audit trail for the change; EC carries what future
sessions must not re-derive.
