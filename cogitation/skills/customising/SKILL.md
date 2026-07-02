---
name: customising
description: Use when the user wants to tailor the cogitation workflow — says "customise", "recustomise", "change the workflow", "make TDD less strict", "turn off finishing-branch", or wants a skill looser or disabled
---

# Customising

Tailor the workflow to this project through a guided negotiation. You advocate
for the core values, offer the lighter mechanism first, then yield — recording
the user's reasoning. Deltas only; upstream files are never edited.

**Announce:** "I'm using the customising skill to tailor the workflow."

## The Loop

```
Listen → Map → Branch on friction → Write deltas → Set gate → Persist decision
```

## Step 0: Load Context

1. Read `cogitation/values.md` (relative to the plugin root) — the six core
   values, which skills serve each, and the pitch for each.
2. Search prior customization decisions:

```
ec_search:
  query: workflow customization
  area: cogitation/customization
  type: decision
```

If EC is down, proceed — the config file is still the source of truth.

## Step 1: Listen

The user describes intent in plain language ("TDD is too strict here", "we
don't do finishing-branch", "make brainstorming advisory"). Restate what you
heard as a concrete change: which skill(s), toggled or re-tuned how.

## Step 2: Map

Map each affected skill to the core value(s) it serves, using `values.md`.
This determines the friction tier.

## Step 3: Branch on Friction

**HARD — continuity-of-memory only.** The one line that never yields. Do not
offer to switch off decision-persistence (`ec_search`/`ec_add` discipline,
`remember`). Say so plainly: "That's the one thing I can't move — without
memory there is no cogitation, just vanilla skills." Offer to serve it
differently (different cadence, different areas), never off.

**Soft — everything else (advocate-then-yield):**
1. **Advocate:** cite the value's pitch from `values.md` and name the concrete
   failure mode the change re-exposes.
2. **Offer the lighter mechanism first:** `rigidity: advisory` before
   `enabled: false`; narrowing a skill's scope before removing it.
3. **Yield:** if the user still wants the stronger change, make it — with an
   explicit recorded acknowledgment ("Understood: disabling finishing-branch;
   you accept that merge hygiene is now manual. Reason: <user's words>").

## Step 4: Write Deltas

Edit `.cogitation/config.json` — **deltas only**, never a copy of defaults:

```json
{
  "workflow": {
    "customized": true,
    "skills": {
      "tdd":              { "rigidity": "advisory" },
      "finishing-branch": { "enabled": false }
    }
  }
}
```

- `skills.<name>.enabled` (bool) — `false` removes it from routing.
- `skills.<name>.rigidity` (`"strict"` | `"advisory"`) — `advisory` means the
  discipline is recommended, not mandatory. Omit for the default (`strict`).
- Remove a delta to restore upstream default — never write defaults out.

## Step 5: Set the Gate

Ensure `workflow.customized` is `true`. This permanently silences the
first-session setup nudge. Set it even when the user reviewed the workflow and
changed nothing — reviewing IS customizing.

## Step 6: Persist the Decision

```
ec_add:
  type: decision
  area: cogitation/customization
  content: [What was changed — skill, delta, and the acknowledgment if yielded]
  rationale: [The user's own reasoning, in their words]
```

The posture takes effect at the next session start (the SessionStart hook
resolves the deltas). Mention that.

## Scope (Slice 1)

Toggle + rigidity only. Adding new skills or overriding upstream ones by
routing (`.claude/skills/` + `route:` remaps) is a future slice — say so if
asked, and record the request in EC so demand is visible.
