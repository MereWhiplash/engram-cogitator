---
name: using-cogitation
description: Use at the start of any conversation — establishes the cogitation workflow and requires invoking the right skill before responding, including before clarifying questions
---

<SUBAGENT-STOP>
If you were dispatched as a subagent to execute a specific task, skip this skill and do the task.
</SUBAGENT-STOP>

<EXTREMELY-IMPORTANT>
If there is even a 1% chance a cogitation skill applies to what you're doing, CHECK for it BEFORE responding — including before clarifying questions. Checking now includes **sizing** (see "Size before routing") — it does not mean brainstorming for everything. If the skill turns out to be wrong, you can drop it. Knowing the concept is not the same as using the skill.
</EXTREMELY-IMPORTANT>

# Using Cogitation

Cogitation is an opinionated development workflow backed by Engram Cogitator (EC) persistent memory. Skills carry the discipline; EC carries the memory across sessions.

## Instruction Priority

1. **User instructions** (CLAUDE.md, direct requests) — always win.
2. **Cogitation skills** — override default behavior where they conflict.
3. **Default system prompt** — lowest.

If CLAUDE.md says "don't use TDD" and a skill says "always," the user wins.

## How to Access Skills

Use the **Skill** tool — it loads the skill content for you to follow. Never `Read` a skill file instead of invoking it. Skills evolve; don't run from memory.

## Size before routing

Before routing any build/change request, size it and **announce the sizing
aloud** — `Sizing: <tier> — <reason>`. The announcement is the license to take
a lighter path; *silent* skipping is the violation.

| Tier | Test | Route |
|---|---|---|
| **TRIVIAL** | no behavior change — typo, rename, comment, config value, doc edit | just do it, verify after |
| **SMALL** | behavior-preserving fix / tiny addition; **two-way door** (easily undone, no new behavior surface, no data/API/schema change) | **tdd** / **debugging** |
| **MEDIUM** | new behavior but contained — one component, no arch/data/API change | **sketching** (half-page combo doc) |
| **FULL** | new behavior surface, one-way door, arch/data/API | **brainstorming** → **writing-plans** |

- Torn between tiers → pick the **heavier**.
- Scope grows mid-flight (one-way door, new dependency, second component) →
  STOP, announce the re-size, **escalate** with context in hand.
- User vetoes a sizing → `ec_add` a learning (`area: workflow/triage`) so
  calibration improves.

## The Workflow

```
sizing → brainstorming → writing-plans → executing-plans → finishing-branch
   │          (FULL)
   └→ sketching (MEDIUM) · tdd/debugging (SMALL) · just do it (TRIVIAL)
                                      │
                  @tdd ·@debugging ·@verifying ·review ladder
```

Pick the entry point by intent:

| You're about to… | Invoke |
|---|---|
| Build a feature / change behavior / "let's add X" | size it first — **brainstorming** (FULL) or **sketching** (MEDIUM) |
| Turn an approved design into tasks | **writing-plans** |
| Implement an existing plan | **executing-plans** (continuous TDD) |
| Write any code | **tdd** (red → green → refactor) |
| Chase a bug | **debugging** |
| Claim something is done | **verifying** |
| Get review on a diff | **requesting-review** (Tier 1) → `@receiving-review` |
| Wrap up / merge | **finishing-branch** |
| Remember a decision/gotcha | **remember** |
| Onboard to a repo | **onboard** · **init** (first-time setup) |
| Tailor the workflow / change how strict it is | **customising** |

**Process skills before implementation skills.** "Build X" → size, then brainstorming or sketching. "Fix bug" → debugging first.

## Workflow posture

When the SessionStart hook injects an "Active workflow posture" block, apply
those per-skill overrides when routing: a **DISABLED** skill is not routed to;
an **advisory** skill's discipline is recommended, not mandatory. User
instructions still win (see Instruction Priority).

## EC memory is part of the loop

Skills search EC (`ec_search`) for prior decisions/patterns/learnings before acting, and store (`ec_add`) durable ones after. If EC isn't running, skills degrade gracefully — proceed without it.

## Graphify (optional, strict when on)

If `.cogitation/config.json` has `graphify.enabled: true`, structural code questions ("what calls X", "where does Y live") go through `graphify query` as a **required** recon step in onboard/brainstorming — not optional. If disabled, skip graphify entirely.

## Red Flags — these thoughts mean STOP, you're rationalizing

| Thought | Reality |
|---|---|
| "This is just a simple question" | Questions are tasks. Check for a skill. |
| "Let me explore the codebase first" | Skills tell you HOW to explore. Check first. |
| "I'll just do this one thing first" | Check BEFORE doing anything. |
| "I'll skip the sizing announcement" | The announcement IS the license. Skipping silently is the violation. |
| "It's *still* small" (as scope grows) | Growth is a re-size. Announce it and escalate. |
| "I remember this skill" | Skills evolve — invoke the current version. |
| "No production code without a failing test… but just this once" | No. @tdd is rigid. |

## Skill Types

**Rigid** (tdd, debugging, verifying): follow exactly — don't adapt away the discipline.
**Flexible** (patterns, onboarding): adapt principles to context. The skill says which it is.
