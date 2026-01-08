//go:build cgo

package storage_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/storage"
	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

func TestSQLiteStorage_Add(t *testing.T) {
	// Use temp file for test database
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	store, err := storage.NewSQLite(f.Name())
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	mem := types.Memory{
		Type:      types.TypeDecision,
		Area:      "auth",
		Content:   "Use JWT tokens",
		Rationale: "Stateless auth",
	}
	embedding := make([]float32, 768)

	result, err := store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if result.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if result.Type != types.TypeDecision {
		t.Errorf("expected type 'decision', got %q", result.Type)
	}
	if !result.IsValid {
		t.Error("expected IsValid to be true")
	}
}

func TestSQLiteStorage_Search(t *testing.T) {
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	store, err := storage.NewSQLite(f.Name())
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Add a memory first
	mem := types.Memory{
		Type:    types.TypeDecision,
		Area:    "auth",
		Content: "Use JWT tokens",
	}
	embedding := make([]float32, 768)
	embedding[0] = 0.5

	_, err = store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Search for it
	results, err := store.Search(ctx, embedding, types.SearchOpts{Limit: 5})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestSQLiteStorage_List(t *testing.T) {
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	store, err := storage.NewSQLite(f.Name())
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Add memories
	for i := 0; i < 3; i++ {
		mem := types.Memory{
			Type:    types.TypeLearning,
			Area:    "api",
			Content: fmt.Sprintf("Learning %d", i),
		}
		embedding := make([]float32, 768)
		_, err = store.Add(ctx, mem, embedding)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	results, err := store.List(ctx, types.ListOpts{Limit: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestSQLiteStorage_Invalidate(t *testing.T) {
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	store, err := storage.NewSQLite(f.Name())
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	mem := types.Memory{
		Type:    types.TypeDecision,
		Area:    "auth",
		Content: "Old decision",
	}
	embedding := make([]float32, 768)

	added, err := store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	err = store.Invalidate(ctx, added.ID, nil)
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	// Should not appear in list (excludes invalid by default)
	results, err := store.List(ctx, types.ListOpts{Limit: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results after invalidation, got %d", len(results))
	}
}

func TestSQLiteStorage_Add_InvalidType(t *testing.T) {
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	store, err := storage.NewSQLite(f.Name())
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	mem := types.Memory{
		Type:    "invalid_type",
		Area:    "auth",
		Content: "Should fail",
	}
	embedding := make([]float32, 768)

	_, err = store.Add(ctx, mem, embedding)
	if err == nil {
		t.Error("expected error for invalid type, got nil")
	}
}
