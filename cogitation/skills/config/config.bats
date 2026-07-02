#!/usr/bin/env bats
# Structural smoke test for the config skill's workflow-manifest docs.
# Run from the repo root: bats cogitation/skills/config/config.bats

@test "config skill documents the workflow manifest section" {
  run cat cogitation/skills/config/SKILL.md
  [ "$status" -eq 0 ]
  [[ "$output" == *"workflow"* ]]
  [[ "$output" == *"rigidity"* ]]
  [[ "$output" == *"customised"* || "$output" == *"customising"* ]]
}
