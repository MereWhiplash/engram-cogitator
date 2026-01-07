package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

func TestMongoDBStorage_Add(t *testing.T) {
	uri := os.Getenv("TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("TEST_MONGODB_URI not set, skipping MongoDB tests")
	}

	ctx := context.Background()
	store, err := storage.NewMongoDB(ctx, uri, "engram_test")
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	mem := storage.Memory{
		Type:        storage.TypeDecision,
		Area:        "auth",
		Content:     "Use JWT tokens",
		Rationale:   "Stateless auth",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Repo:        "testorg/testrepo",
	}
	embedding := make([]float32, 768)
	embedding[0] = 0.5

	result, err := store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if result.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if result.AuthorEmail != "test@example.com" {
		t.Errorf("expected author email 'test@example.com', got %q", result.AuthorEmail)
	}
}

func TestMongoDBStorage_Search(t *testing.T) {
	uri := os.Getenv("TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("TEST_MONGODB_URI not set, skipping MongoDB tests")
	}

	ctx := context.Background()
	store, err := storage.NewMongoDB(ctx, uri, "engram_test")
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	mem := storage.Memory{
		Type:        storage.TypeLearning,
		Area:        "api",
		Content:     "Rate limiting patterns",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Repo:        "testorg/testrepo",
	}
	embedding := make([]float32, 768)
	embedding[0] = 0.8

	_, err = store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	results, err := store.Search(ctx, embedding, storage.SearchOpts{Limit: 5})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) < 1 {
		t.Error("expected at least 1 result")
	}
}

func TestMongoDBStorage_List(t *testing.T) {
	uri := os.Getenv("TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("TEST_MONGODB_URI not set, skipping MongoDB tests")
	}

	ctx := context.Background()
	store, err := storage.NewMongoDB(ctx, uri, "engram_test")
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	mem := storage.Memory{
		Type:        storage.TypePattern,
		Area:        "testing",
		Content:     "Always use table-driven tests",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Repo:        "testorg/testrepo",
	}
	embedding := make([]float32, 768)

	_, err = store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	results, err := store.List(ctx, storage.ListOpts{Limit: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(results) < 1 {
		t.Error("expected at least 1 result")
	}
}
