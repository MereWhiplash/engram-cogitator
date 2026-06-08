package main

import "testing"

func TestBuildStorageConfig_SQLite(t *testing.T) {
	cfg, err := buildStorageConfig("sqlite", "/tmp/ec/memory.db", "", "", "engram")
	if err != nil {
		t.Fatalf("sqlite should be allowed for api: %v", err)
	}
	if cfg.Driver != "sqlite" || cfg.SQLitePath != "/tmp/ec/memory.db" {
		t.Fatalf("got driver=%q path=%q", cfg.Driver, cfg.SQLitePath)
	}
}

func TestBuildStorageConfig_SQLiteRequiresPath(t *testing.T) {
	if _, err := buildStorageConfig("sqlite", "", "", "", "engram"); err == nil {
		t.Fatal("expected error when sqlite db-path is empty")
	}
}

func TestBuildStorageConfig_Postgres(t *testing.T) {
	cfg, err := buildStorageConfig("postgres", "", "postgres://x", "", "engram")
	if err != nil || cfg.PostgresDSN != "postgres://x" {
		t.Fatalf("postgres config broken: %v cfg=%+v", err, cfg)
	}
}
