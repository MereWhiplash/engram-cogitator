#!/usr/bin/env bats
# Structural smoke test for the customising skill.
# Run from the repo root: bats cogitation/skills/customising/customising.bats

@test "customising skill has frontmatter, triggers, advocate-yield loop, hard memory line" {
  run cat cogitation/skills/customising/SKILL.md
  [ "$status" -eq 0 ]
  [[ "$output" == *"name: customising"* ]]
  [[ "$output" == *"customise"* ]]
  [[ "$output" == *"recustomise"* ]]
  [[ "$output" == *"values.md"* ]]             # cites the taxonomy
  [[ "$output" == *"lighter mechanism"* ]]     # offers softer option first
  [[ "$output" == *"acknowledgment"* ]]        # records the yield
  [[ "$output" == *"ec_add"* ]]                # persists the decision
  [[ "$output" == *"customized"* ]]            # sets the gate
  [[ "$output" == *"never"* ]]                 # memory never yields
}
