---
name: brainstorming
description: Use before any feature work or new feature — interviews you about the feature one question at a time, pushes back on gaps, and saves a design doc with a decisions log
---

# Brainstorming

Turn ideas into designs through a rigorous, collaborative interview. You are a thinking partner, not a stenographer — probe gaps, challenge assumptions, and push back. Stay conversational; ask, don't interrogate.

**Announce:** "I'm using the brainstorming skill to explore this feature."

## The Flow

```
Recon → Branch Setup → Interview (with pushback) → Approaches → Design → Save
```

## Step 1: Recon (EC + codebase)

Before questioning, gather context so your questions are sharp:

```
ec_search:
  query: [feature area]
  type: decision

ec_search:
  query: [related patterns]
  type: pattern

ec_search:
  query: [similar features]
```

Also skim the codebase for the relevant tech stack, conventions, and dependencies. Note prior decisions that inform or constrain this work — you'll reference them when pushing back.

**Graphify (strict when enabled):** Read `.cogitation/config.json`. If `graphify.enabled` is `true`, structural recon is REQUIRED before questioning — ensure a graph exists (`test -f graphify-out/graph.json || graphify update .`, no model needed) then run `graphify query "how does <feature area> work today?"` to ground your questions in the real call-graph. If absent/`false`, skip.

## Step 2: Branch Setup

```bash
git branch --show-current
git status
```

Load branch convention:
```
ec_search:
  query: project config
  type: config
```

**REQUIRED: Use the AskUserQuestion tool** to determine branching:

**If on main:**
```json
{ "questions": [{ "question": "Create a feature branch for this work?", "header": "Branch",
  "options": [ { "label": "Yes", "description": "Create {convention}/<name> branch" },
               { "label": "No", "description": "Stay on main" } ], "multiSelect": false }] }
```

**If on a feature branch:**
```json
{ "questions": [{ "question": "Where should this new work branch from?", "header": "Branch",
  "options": [ { "label": "Current branch", "description": "Nested feature off current work" },
               { "label": "Main", "description": "Fresh start from main" } ], "multiSelect": false }] }
```

Features can spawn sub-features.

## Step 3: Interview — one question at a time

**REQUIRED: Use the AskUserQuestion tool.** Ask **one question at a time** with **2–4 non-obvious options** (lead with your recommended option). Users can always type free text.

Walk the coverage areas, splitting deeper as needed:

- **Problem** — what are we actually solving? What happens if we don't?
- **Users** — who uses this, with what permissions?
- **Technical approach** — where does it live, what does it touch?
- **Risks & edge cases** — failure modes, empty/error/concurrent states
- **Constraints** — scope, performance, compatibility
- Then as relevant: **API design · data model · state · migration**

Track coverage; keep going until each discovered area is resolved. Don't ask the obvious — ask what reveals hidden assumptions.

### Active pushback (do this, don't just record)

Push back immediately when you hit:
- **Contradiction** with an earlier answer or an EC decision → name it, cite the memory, ask which wins
- **Over-engineering** for the stated scope → propose the simpler path (YAGNI)
- **Missing edge case** → surface it as a question, don't let it slide

> "That contradicts what you said about X — and EC has a decision [area] saying Y. Which holds?"

### Security HARD-BLOCKS (halt the interview)

If the feature involves any of these **unaddressed**, STOP and require an explicit decision before continuing or writing the design:
- PII stored/transmitted without encryption
- Auth bypass or missing authorization checks
- Injection risk (SQL/command/template) on untrusted input
- Plaintext secrets / credentials
- No rate limiting on an abusable endpoint
- No data deletion / retention strategy where data is collected

Flagged items must appear in the design's Security section regardless of resolution.

## Step 4: Explore Approaches

Propose 2–3 approaches with trade-offs; lead with your recommendation and why. Search EC for prior decisions on anything you're weighing:

```
ec_search:
  query: [technology or pattern being considered]
  type: decision
```

> "I'd go with A because [reason]. B works if [condition]. C is overkill unless [edge case]."

## Step 5: Present Design

Present in chunks (200–300 words), checking after each. Cover architecture/components, data flow, key implementation details, edge cases/error handling, and security. Backtrack if something doesn't fit.

## Step 6: Save Design

Write to `docs/designs/YYYY-MM-DD-<topic>.md`, including:
- Summary of what we're building; goals / non-goals
- Architecture, data flow, key decisions and rationale
- **Security Considerations** (every hard-block item + resolution)
- **Decisions Log** — audit trail of pushback, contradictions raised, and what was decided (the disagreements, not just the conclusions)
- **Implementation Order** — tasks in dependency order (feeds writing-plans)
- EC memories consulted; placeholder: `See: docs/plans/YYYY-MM-DD-<topic>.md`

## Step 7: Store Decisions

For each significant architectural decision:

```
ec_add:
  type: decision
  area: [component]
  content: [What was decided and why]
  rationale: [Trade-offs considered]
```

## Handoff

> "Design saved to `docs/designs/YYYY-MM-DD-<topic>.md`. Ready to create the implementation plan?"

If yes → **Use @writing-plans** (all tasks follow **@tdd** red-green-refactor).
