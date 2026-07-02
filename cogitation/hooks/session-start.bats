#!/usr/bin/env bats
# Tests for the SessionStart hook's workflow-config plumbing.
# Run from the repo root: bats cogitation/hooks/session-start.bats
# EC_COG_CONFIG points the hook at a fixture config for side-effect-free runs.

setup() { TMP="$(mktemp -d)"; }
teardown() { rm -rf "$TMP"; }

@test "no config: injects the default-profile nudge" {
  EC_COG_CONFIG="$TMP/missing.json" run bash cogitation/hooks/session-start.sh
  [ "$status" -eq 0 ]
  [[ "$output" == *"customise"* ]]
  [[ "$output" == *"default cogitation profile"* ]]
}

@test "customized=true: no nudge" {
  printf '{"workflow":{"customized":true}}' > "$TMP/c.json"
  EC_COG_CONFIG="$TMP/c.json" run bash cogitation/hooks/session-start.sh
  [ "$status" -eq 0 ]
  [[ "$output" != *"default cogitation profile"* ]]
}
