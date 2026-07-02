#!/usr/bin/env bats
# Structural smoke test for the cog-init skill.
# Run from the repo root: bats cogitation/skills/cog-init/cog-init.bats

@test "cog-init references the workflow gate and customising handoff" {
  run cat cogitation/skills/cog-init/SKILL.md
  [ "$status" -eq 0 ]
  [[ "$output" == *"workflow"* ]]
  [[ "$output" == *"customising"* ]]
}
