#!/usr/bin/env bats
# Structural smoke test for the sketching skill (MEDIUM-tier path).
# Run from the repo root: bats cogitation/skills/sketching/sketching.bats

@test "sketching skill has frontmatter, combo doc, discipline, escalation" {
  run cat cogitation/skills/sketching/SKILL.md
  [ "$status" -eq 0 ]
  [[ "$output" == *"name: sketching"* ]]
  [[ "$output" == *"docs/sketches/"* ]]        # the combo doc home
  [[ "$output" == *"half-page"* ]]             # size cap
  [[ "$output" == *"rejected"* ]]              # approach + rejected alternative
  [[ "$output" == *"checklist"* ]]             # 2–5 task checklist
  [[ "$output" == *"tdd"* ]]                   # same execution discipline
  [[ "$output" == *"escalate"* ]]              # one-way door → brainstorming
  [[ "$output" == *"ec_add"* ]]                # persists decisions
}
