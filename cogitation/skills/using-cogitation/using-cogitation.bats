#!/usr/bin/env bats
# Structural smoke test for the using-cogitation dispatcher.
# Run from the repo root: bats cogitation/skills/using-cogitation/using-cogitation.bats

@test "dispatcher documents posture-application and the customise path" {
  run cat cogitation/skills/using-cogitation/SKILL.md
  [ "$status" -eq 0 ]
  [[ "$output" == *"posture"* ]]          # applies injected posture
  [[ "$output" == *"customising"* ]]      # routes the customise intent
}

@test "dispatcher documents the four-tier sizing triage" {
  run cat cogitation/skills/using-cogitation/SKILL.md
  [ "$status" -eq 0 ]
  [[ "$output" == *"Sizing:"* ]]          # the announced-sizing format
  [[ "$output" == *"TRIVIAL"* ]]
  [[ "$output" == *"SMALL"* ]]
  [[ "$output" == *"MEDIUM"* ]]
  [[ "$output" == *"FULL"* ]]
  [[ "$output" == *"sketching"* ]]        # MEDIUM routes to sketching
  [[ "$output" == *"two-way door"* ]]     # the sizing test
  [[ "$output" == *"heavier"* ]]          # torn between tiers → heavier
  [[ "$output" == *"escalate"* ]]         # mid-flight re-size
  [[ "$output" != *"too simple to need a design"* ]]  # old red flag deleted
}
