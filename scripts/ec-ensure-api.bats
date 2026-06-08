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
