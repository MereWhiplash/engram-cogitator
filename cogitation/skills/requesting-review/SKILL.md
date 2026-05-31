---
name: requesting-review
description: Use when a diff is large or high-risk and Tier 0 inline self-review isn't enough — dispatches an independent reviewer subagent
---

# Requesting Code Review

Dispatch an independent reviewer for changes too large or risky for inline self-review. The reviewer gets precisely crafted context — never your session history — so it judges the work product, not your reasoning, and your own context stays clean.

**Announce:** "I'm using the requesting-review skill to dispatch an independent review."

**Where this sits:** Tier 1 of the review ladder (see `@executing-plans` Step 5). Default to **Tier 0** inline self-review for small/medium changes. Escalate to **Tier 2** external/adversarial review when you want a second model — only if `codex` is enabled in `.cogitation/config.json`:
- Code: `/codex:adversarial-review --background` (official `codex-plugin-cc`)
- Design/plan docs: `@codex-review`

## Step 1: Verify Ready @verifying

```bash
{test_command}
{lint_command}
{build_command}
```

**If failures:** Stop. Fix first using `@debugging`. Never request review on red.

## Step 2: Load Review Context (EC)

Search EC so the reviewer gets codebase-specific signal:

```
ec_search:
  query: code review feedback
  type: learning

ec_search:
  query: review pattern
  type: pattern
```

Note prior feedback patterns to fold into the dispatch context.

## Step 3: Dispatch the Reviewer

Get the range:
```bash
BASE_SHA=$(git rev-parse origin/main)   # or HEAD~N for a single task
HEAD_SHA=$(git rev-parse HEAD)
```

Dispatch the **`general-purpose`** agent (no named agent — one source of truth) with this template:

```
Task tool (general-purpose):
  description: "Review code changes"
  prompt: |
    You are a Senior Code Reviewer. Review completed work against its plan/requirements
    and identify issues before they cascade. Read the actual diff — don't trust summaries.

    ## What Was Implemented
    {DESCRIPTION}

    ## Requirements / Plan
    {PLAN_OR_REQUIREMENTS}     (e.g. docs/plans/YYYY-MM-DD-<topic>.md)

    ## Git Range
    Base: {BASE_SHA}
    Head: {HEAD_SHA}
    ```bash
    git diff --stat {BASE_SHA}..{HEAD_SHA}
    git diff {BASE_SHA}..{HEAD_SHA}
    ```

    ## EC Context (codebase conventions / prior decisions)
    {EC_CONTEXT}

    ## What to Check
    - Plan alignment: all planned functionality present; deviations justified, not accidental
    - Code quality: separation of concerns, error handling, types, DRY without premature abstraction, edge cases
    - Architecture: sound decisions, security, integrates cleanly with surrounding code
    - Testing: tests assert real behavior (not mocks); edge cases; all passing

    ## Calibration
    Categorize by ACTUAL severity — not everything is Critical. Only flag issues that
    would cause real problems during implementation or in production. Skip style,
    wording, and formatting preferences. Acknowledge what was done well first so the
    rest of the feedback is trusted.

    ## Output
    ### Strengths
    ### Issues
    #### Critical (Must Fix)    — bugs, security, data loss, broken behavior
    #### Important (Should Fix) — architecture, missing features, test gaps
    #### Minor (Nice to Have)   — style, optimization, docs
    (each: file:line · what's wrong · why it matters · how to fix)
    ### Assessment — Ready to merge? [Yes | No | With fixes] + 1-2 sentence reason
```

Placeholders: `{DESCRIPTION}` what you built · `{PLAN_OR_REQUIREMENTS}` plan path/task text · `{BASE_SHA}`/`{HEAD_SHA}` · `{EC_CONTEXT}` relevant decisions/patterns from Step 2.

## Step 4: Act on Feedback

- Fix **Critical** and **Important** before proceeding.
- Note **Minor** for later.
- Push back (with technical reasoning) if the reviewer is wrong — process with `@receiving-review`.

**If approved:**
> "Review passed. Ready to finish the branch?"

If yes → **Use @finishing-branch**

## Step 5: Store Review Learnings (EC)

If review reveals a durable pattern worth remembering:

```
ec_add:
  type: learning
  area: code-review
  content: [What reviewers commonly catch in this codebase]
  rationale: Recurring review feedback
```

## What Makes a Good Review Request

| Do | Don't |
|----|----|
| Verify green before requesting | Request review on failing tests |
| Give the reviewer the diff range + plan | Dump code without context |
| Pass EC conventions/decisions | Assume the reviewer knows the codebase |
| Keep scope focused | Request review of WIP code |
