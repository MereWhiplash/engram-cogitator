#!/usr/bin/env bats

@test "ensure-api script exists and is executable" {
  [ -x scripts/ec-ensure-api.sh ]
}

@test "passes shellcheck" {
  run shellcheck scripts/ec-ensure-api.sh
  [ "$status" -eq 0 ]
}

@test "dry-run prints docker run with loopback bind and restart policy" {
  EC_DRY_RUN=1 run bash scripts/ec-ensure-api.sh
  [[ "$output" == *"127.0.0.1:8080:8080"* ]]
  [[ "$output" == *"--restart unless-stopped"* ]]
  [[ "$output" == *"--name engram-cogitator-api"* ]]
  [[ "$output" == *"--storage-driver sqlite"* ]]
}

@test "dry-run prints ollama ensure with volume and restart policy" {
  EC_DRY_RUN=1 run bash scripts/ec-ensure-api.sh
  [[ "$output" == *"--name engram-ollama"* ]]
  [[ "$output" == *"ollama_data:/root/.ollama"* ]]
  [[ "$output" == *"ollama/ollama"* ]]
  # both containers get a restart policy — an unrestarted embedder 500s the API
  ollama_line="$(printf '%s\n' "$output" | grep -- '--name engram-ollama')"
  [[ "$ollama_line" == *"--restart unless-stopped"* ]]
}
