package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

func TestNew_SQLite(t *testing.T) {
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	ctx := context.Background()
	store, err := storage.New(ctx, storage.Config{
		Driver:     "sqlite",
		SQLitePath: f.Name(),
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	// Verify it works
	mem := storage.Memory{
		Type:    storage.TypeDecision,
		Area:    "test",
		Content: "Test content",
	}
	_, err = store.Add(ctx, mem, make([]float32, 768))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
}

func TestNew_UnknownDriver(t *testing.T) {
	ctx := context.Background()
	_, err := storage.New(ctx, storage.Config{
		Driver: "unknown",
	})
	if err == nil {
		t.Error("expected error for unknown driver")
	}
}

func TestNew_SQLite_MissingPath(t *testing.T) {
	ctx := context.Background()
	_, err := storage.New(ctx, storage.Config{
		Driver: "sqlite",
	})
	if err == nil {
		t.Error("expected error for missing sqlite path")
	}
}

func TestNew_Postgres_MissingDSN(t *testing.T) {
	ctx := context.Background()
	_, err := storage.New(ctx, storage.Config{
		Driver: "postgres",
	})
	if err == nil {
		t.Error("expected error for missing postgres DSN")
	}
}

func TestNew_MongoDB_MissingURI(t *testing.T) {
	ctx := context.Background()
	_, err := storage.New(ctx, storage.Config{
		Driver: "mongodb",
	})
	if err == nil {
		t.Error("expected error for missing mongodb URI")
	}
}
