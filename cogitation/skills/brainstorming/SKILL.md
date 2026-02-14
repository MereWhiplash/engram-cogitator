---
name: brainstorming
description: Explores requirements, searches past decisions, creates design documents, and sets up feature branches. Use before any feature work or when starting a new feature.
---

# Brainstorming

Turn ideas into designs through collaborative dialogue.

**Announce:** "I'm using the brainstorming skill to explore this feature."

## The Flow

```
EC Search → Branch Setup → Q&A → Approaches → Design → Save
```

## Step 1: Search Memory

Before designing, check what we've done before:

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

Note relevant past decisions that might inform or constrain the design.

## Step 2: Branch Setup

Check current state:

```bash
git branch --show-current
git status
```

Load branch convention from config:

```
ec_search:
  query: project config
  type: config
```

**REQUIRED: Use the AskUserQuestion tool** to determine branching:

**If on main:**
```json
{
  "questions": [{
    "question": "Create a feature branch for this work?",
    "header": "Branch",
    "options": [
      { "label": "Yes", "description": "Create {convention}/<name> branch" },
      { "label": "No", "description": "Stay on main" }
    ],
    "multiSelect": false
  }]
}
```

**If on feature branch:**
```json
{
  "questions": [{
    "question": "Where should this new work branch from?",
    "header": "Branch",
    "options": [
      { "label": "Current branch", "description": "Nested feature off current work" },
      { "label": "Main", "description": "Fresh start from main" }
    ],
    "multiSelect": false
  }]
}
```

This enables branching cycles - features can spawn sub-features.

## Step 3: Understand the Idea

**REQUIRED: Use the AskUserQuestion tool** for all clarifying questions.

- 1-4 questions per tool call
- 2-4 options per question
- Users can always type free text (implicit "Other" option)
- Use `multiSelect: true` when multiple options can apply

Focus on:
- What problem are we solving?
- What are the constraints?
- What does success look like?

Example:
```json
{
  "questions": [{
    "question": "Who should have access to this feature?",
    "header": "Access",
    "options": [
      { "label": "All users", "description": "Any authenticated user" },
      { "label": "Admins only", "description": "Requires admin role" },
      { "label": "Role-based", "description": "Configurable per role" }
    ],
    "multiSelect": false
  }]
}
```

## Step 4: Explore Approaches

Propose 2-3 approaches with trade-offs. Lead with your recommendation and why.

Search EC for relevant prior decisions:

```
ec_search:
  query: [technology or pattern being considered]
  type: decision
```

Keep it conversational:
> "I'd go with Option A because [reason]. Option B would work if [condition]. Option C is overkill unless [edge case]."

## Step 5: Present Design

Once you understand what we're building, present the design in chunks (200-300 words each). Check after each section if it looks right.

Cover:
- Architecture / component structure
- Data flow
- Key implementation details
- Edge cases / error handling

Be ready to backtrack if something doesn't fit.

## Step 6: Save Design

Write to `docs/designs/YYYY-MM-DD-<topic>.md`

Include:
- Summary of what we're building
- Key decisions and rationale
- Reference to any EC memories consulted
- Placeholder for implementation plan: `See: docs/plans/YYYY-MM-DD-<topic>.md`

## Step 7: Store Decisions

For each significant architectural decision made:

```
ec_add:
  type: decision
  area: [component]
  content: [What was decided and why]
  rationale: [Trade-offs considered]
```

## Handoff

After saving:
> "Design saved to `docs/designs/YYYY-MM-DD-<topic>.md`. Ready to create the implementation plan?"

If yes → **Use @writing-plans**
