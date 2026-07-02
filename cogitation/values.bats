#!/usr/bin/env bats
# Structural smoke test for the core-values taxonomy.
# Run from the repo root: bats cogitation/values.bats

@test "values.md declares all six core values with friction tiers" {
  run cat cogitation/values.md
  [ "$status" -eq 0 ]
  for v in "continuity-of-memory" "think-before-building" "prove-before-claiming" \
           "understand-before-changing" "keep-it-simple" "small-incremental-prs"; do
    [[ "$output" == *"$v"* ]]
  done
  [[ "$output" == *"HARD"* ]]   # memory is the hard line
  [[ "$output" == *"advocate-then-yield"* ]]
}
