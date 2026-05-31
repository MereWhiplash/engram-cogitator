---
name: codex-review
description: Use when you want an independent second-opinion review of a design doc or implementation plan from OpenAI Codex before executing — only if codex is enabled in .cogitation/config.json
---

# Codex Review (Tier 2 — design/plan docs)

Get an independent second opinion from OpenAI Codex on a **design or plan document** before committing to execution.

**Announce:** "I'm using the codex-review skill to get a second opinion from Codex."

## Scope

- **This skill = document review** (design docs, implementation plans). The official `codex-plugin-cc` reviews *diffs*, not arbitrary docs — this fills that gap.
- For **code / adversarial review of a diff**, prefer the official plugin instead: `/codex:adversarial-review --background`.

## Step 0: Gate

Only run if **`.cogitation/config.json` has `codex.enabled: true`** AND `codex` is on PATH:

```bash
command -v codex >/dev/null || { echo "codex not installed; skipping"; exit 0; }
```

If disabled or missing, say so and fall back to Tier 0/1 review. Don't install anything.

## Step 1: Identify the Artifact

If unspecified, pick the most recent:
```bash
ls -t docs/designs/*.md docs/plans/*.md 2>/dev/null | head -1
```
If both a design and a plan are plausible, ask with AskUserQuestion which to review.

## Step 2: Build the Review Prompt

**Design docs** — focus: feasibility vs. the actual codebase, missed edge cases, simpler alternatives, data-model/migration risk, security & permissions.

**Implementation plans** — focus: task ordering/dependencies, completeness vs. design, test coverage of described behavior, file paths exist, scope.

Output structure to request: `CONCERNS` (block execution) · `SUGGESTIONS` (non-blocking) · `STRENGTHS` (brief).

## Step 3: Run Codex — NON-BLOCKING

**Run Codex in the background and let the harness notify you on completion. Do NOT block the session in the foreground.**

Write stdout and stderr to *separate* files (don't merge — stderr is progress/log noise):

```bash
mkdir -p .cogitation/codex
OUT=.cogitation/codex/review-$(git rev-parse --short HEAD).md
ERR=.cogitation/codex/review.log
cat <artifact-path> | codex -a never exec \
  --model gpt-5.4 \
  --sandbox read-only \
  --ephemeral \
  --color never \
  -C "$(git rev-parse --show-toplevel)" \
  "<review-prompt>" \
  > "$OUT" 2> "$ERR"
```

Dispatch this with the **Bash tool using `run_in_background: true`**. You'll get a completion notification — then do **ONE** `Read` of `$OUT` (the clean stdout review). Never tail/cat/grep-loop the file while it grows; that re-pulls overlapping content into context and wastes tokens — the very thing that made the old foreground version janky.

**Flags:** `-a never` (before `exec`) non-interactive · `--model gpt-5.4` pinned · `--sandbox read-only` (review, never edits) · `--ephemeral` no session state · `--color never` clean capture · `-C <root>` repo context. Never `--full-auto` (grants write). If `codex` isn't found, stop — don't auto-install.

## Step 4: Present Review

Show it with attribution, then ask how to proceed (Address concerns / Note and proceed / Discuss) via AskUserQuestion.

> **Codex Review (gpt-5.4):**
> [contents of $OUT]

## Step 5: Apply (if chosen)

Edit the design/plan to address CONCERNS, then note it in the artifact:
```markdown
**Codex Review:** YYYY-MM-DD. Addressed: [list]. Deferred: [list].
```
Store anything durable with `@remember` (EC).

## Handoff

- Reviewed a design → "Ready to create the implementation plan?" → `@writing-plans`
- Reviewed a plan → "Ready to execute?" → `@executing-plans`
