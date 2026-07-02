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

@test "advisory rigidity emits a posture line; strict does not" {
  printf '{"workflow":{"customized":true,"skills":{"tdd":{"rigidity":"advisory"},"verifying":{"rigidity":"strict"}}}}' > "$TMP/c.json"
  EC_COG_CONFIG="$TMP/c.json" run bash cogitation/hooks/session-start.sh
  [ "$status" -eq 0 ]
  [[ "$output" == *"\`tdd\` rigidity is **advisory**"* ]]
  [[ "$output" != *"\`verifying\` rigidity"* ]]   # strict is the default → no line
}

@test "enabled:false emits a DISABLED posture line" {
  printf '{"workflow":{"customized":true,"skills":{"finishing-branch":{"enabled":false}}}}' > "$TMP/c.json"
  EC_COG_CONFIG="$TMP/c.json" run bash cogitation/hooks/session-start.sh
  [ "$status" -eq 0 ]
  [[ "$output" == *"\`finishing-branch\` is DISABLED"* ]]
}

@test "malformed config: treated as uncustomized, no crash" {
  printf 'not json {{{' > "$TMP/bad.json"
  EC_COG_CONFIG="$TMP/bad.json" run bash cogitation/hooks/session-start.sh
  [ "$status" -eq 0 ]
  [[ "$output" == *"default cogitation profile"* ]]
}

@test "wrong-typed but valid JSON: degrades, no crash" {
  printf '{"workflow":"nope"}' > "$TMP/c.json"
  EC_COG_CONFIG="$TMP/c.json" run bash cogitation/hooks/session-start.sh
  [ "$status" -eq 0 ]
  [[ "$output" == *"default cogitation profile"* ]]

  printf '{"workflow":{"customized":true,"skills":"nope"}}' > "$TMP/c.json"
  EC_COG_CONFIG="$TMP/c.json" run bash cogitation/hooks/session-start.sh
  [ "$status" -eq 0 ]
  [[ "$output" == *"hookSpecificOutput"* ]]
}

@test "hook passes shellcheck" {
  run shellcheck cogitation/hooks/session-start.sh
  [ "$status" -eq 0 ]
}
