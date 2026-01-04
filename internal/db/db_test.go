package db

import (
	"testing"
)

// mockEmbedding returns a deterministic 768-dim embedding for testing
func mockEmbedding(seed float32) []float32 {
	emb := make([]float32, 768)
	for i := range emb {
		emb[i] = seed + float32(i)*0.001
	}
	return emb
}

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestAdd(t *testing.T) {
	db := setupTestDB(t)

	mem, err := db.Add("decision", "testing", "Use mocks for unit tests", "Faster and deterministic", mockEmbedding(1.0))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if mem.ID != 1 {
		t.Errorf("expected ID 1, got %d", mem.ID)
	}
	if mem.Type != "decision" {
		t.Errorf("expected type decision, got %s", mem.Type)
	}
	if mem.Area != "testing" {
		t.Errorf("expected area testing, got %s", mem.Area)
	}
	if mem.Content != "Use mocks for unit tests" {
		t.Errorf("unexpected content: %s", mem.Content)
	}
	if !mem.IsValid {
		t.Error("expected IsValid to be true")
	}
}

func TestList(t *testing.T) {
	db := setupTestDB(t)

	// Add 3 memories
	db.Add("decision", "api", "Use REST", "", mockEmbedding(1.0))
	db.Add("learning", "api", "Rate limiting needed", "", mockEmbedding(2.0))
	db.Add("pattern", "ui", "Component composition", "", mockEmbedding(3.0))

	// List all
	mems, err := db.List(10, "", "", false)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(mems) != 3 {
		t.Errorf("expected 3 memories, got %d", len(mems))
	}

	// List by type
	mems, err = db.List(10, "decision", "", false)
	if err != nil {
		t.Fatalf("List by type failed: %v", err)
	}
	if len(mems) != 1 {
		t.Errorf("expected 1 decision, got %d", len(mems))
	}

	// List by area
	mems, err = db.List(10, "", "api", false)
	if err != nil {
		t.Fatalf("List by area failed: %v", err)
	}
	if len(mems) != 2 {
		t.Errorf("expected 2 api memories, got %d", len(mems))
	}
}

func TestInvalidate(t *testing.T) {
	db := setupTestDB(t)

	mem, _ := db.Add("decision", "test", "Old decision", "", mockEmbedding(1.0))
	newMem, _ := db.Add("decision", "test", "New decision", "", mockEmbedding(2.0))

	// Invalidate with superseded_by
	err := db.Invalidate(mem.ID, &newMem.ID)
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	// Should not appear in default list
	mems, _ := db.List(10, "", "", false)
	if len(mems) != 1 {
		t.Errorf("expected 1 valid memory, got %d", len(mems))
	}

	// Should appear with includeInvalid
	mems, _ = db.List(10, "", "", true)
	if len(mems) != 2 {
		t.Errorf("expected 2 memories with invalid, got %d", len(mems))
	}
}

func TestSearch(t *testing.T) {
	db := setupTestDB(t)

	// Add memories with different embeddings
	db.Add("decision", "docker", "Use Debian for builds", "", mockEmbedding(1.0))
	db.Add("learning", "go", "Use interfaces", "", mockEmbedding(5.0))
	db.Add("pattern", "testing", "Table driven tests", "", mockEmbedding(10.0))

	// Search with embedding close to first memory
	results, err := db.Search(mockEmbedding(1.1), 3, "", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// First result should be closest to search embedding
	if results[0].Area != "docker" {
		t.Errorf("expected docker first (closest), got %s", results[0].Area)
	}
}
