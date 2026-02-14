---
name: writing-skills
description: Creates new cogitation skills following best practices and EC integration patterns. Use when adding new capabilities or workflows to the skill set.
---

# Writing Skills

Create skills that integrate with EC and follow cogitation conventions.

**Announce:** "I'm using the writing-skills skill to create a new skill."

## The Structure

Every skill needs a `SKILL.md` file in `cogitation/skills/<skill-name>/`:

```yaml
---
name: skill-name
description: [Third person verb] what it does. Mentions EC integration. Use when [trigger conditions].
---

# Skill Title

Brief description.

**Announce:** "I'm using the skill-name skill to [action]."

## Step 0: Load Context

ec_search for relevant config/patterns/learnings.

## Instructions
[Clear, step-by-step guidance with EC integration]

## Store Learnings
[When and how to store memories]

## References
[Link to other skills with @notation]
```

## Naming Conventions

- **Gerund form preferred**: `debugging`, `verifying`, `writing-plans`
- **Lowercase with hyphens**: `requesting-review`, not `RequestingReview`
- **Max 64 characters**
- **No reserved words**: `anthropic`, `claude`

## Description Rules

1. **Third person** - "Provides...", "Creates...", "Handles..."
2. **What + When** - What it does AND when to use it
3. **Mention EC** - Note the EC integration
4. **Include trigger words** - Keywords users would naturally say
5. **Max 1024 characters**

**Good:**
```yaml
description: Provides systematic debugging workflow that finds root cause before fixing. Searches EC for prior issues and stores learnings. Use when encountering bugs, test failures, or unexpected behavior.
```

**Bad:**
```yaml
description: Use this for debugging
```

## Content Guidelines

- **Under 500 lines** - Split into separate files if longer
- **Concise** - Claude is smart, don't over-explain
- **Use @notation** - Reference other skills: `@tdd`, `@verifying`
- **Include announcement** - "I'm using the X skill to..."
- **Provide examples** - Concrete, not abstract

## EC Integration Requirements

Every cogitation skill MUST integrate with EC. Include:

### Step 0: Load Context

```markdown
## Step 0: Load Context

Get project config and relevant memories:

\`\`\`
ec_search:
  query: project config
  type: config

ec_search:
  query: [relevant area]
  type: pattern|learning|decision
\`\`\`
```

### Store Learnings Section

```markdown
## Store Learnings

If [condition worth storing]:

\`\`\`
ec_add:
  type: learning|pattern|decision
  area: [component]
  content: [What to remember]
  rationale: [Why it matters]
\`\`\`
```

## Skill References

Use `@skill-name` to reference other skills:
- In headings: `## Phase 4: Fix @tdd`
- In text: `Use @verifying before claiming success`
- In handoffs: `If yes → **Use @finishing-branch**`

## Progressive Disclosure

For complex skills, use separate files:

```
my-skill/
├── SKILL.md          # Overview (< 500 lines)
├── reference.md      # Detailed docs (loaded when needed)
└── examples.md       # Usage examples
```

Reference in SKILL.md:
```markdown
For detailed API reference, see [reference.md](reference.md)
```

## After Creating

1. Test with a real task
2. Verify EC integration works
3. Check @references resolve
4. Iterate based on usage

## Store New Skill Pattern

After creating a new skill:

```
ec_add:
  type: pattern
  area: cogitation
  content: Created skill [name] for [purpose]. Key pattern: [main workflow].
  rationale: Extends cogitation capabilities
```
