package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

func TestPostgresStorage_Add(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set, skipping Postgres tests")
	}

	ctx := context.Background()
	store, err := storage.NewPostgres(ctx, dsn)
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

func TestPostgresStorage_Search(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set, skipping Postgres tests")
	}

	ctx := context.Background()
	store, err := storage.NewPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	// Add a memory
	mem := storage.Memory{
		Type:        storage.TypeDecision,
		Area:        "auth",
		Content:     "Use JWT tokens",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Repo:        "testorg/testrepo",
	}
	embedding := make([]float32, 768)
	embedding[0] = 0.5

	_, err = store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Search - should find it
	results, err := store.Search(ctx, embedding, storage.SearchOpts{Limit: 5})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) < 1 {
		t.Error("expected at least 1 result")
	}
}

func TestPostgresStorage_SearchByRepo(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set, skipping Postgres tests")
	}

	ctx := context.Background()
	store, err := storage.NewPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	embedding := make([]float32, 768)

	// Add memories to different repos
	mem1 := storage.Memory{
		Type:        storage.TypeDecision,
		Area:        "auth",
		Content:     "Repo A decision",
		AuthorName:  "User A",
		AuthorEmail: "a@example.com",
		Repo:        "org/repo-a",
	}
	mem2 := storage.Memory{
		Type:        storage.TypeDecision,
		Area:        "auth",
		Content:     "Repo B decision",
		AuthorName:  "User B",
		AuthorEmail: "b@example.com",
		Repo:        "org/repo-b",
	}

	_, _ = store.Add(ctx, mem1, embedding)
	_, _ = store.Add(ctx, mem2, embedding)

	// Search scoped to repo-a
	results, err := store.Search(ctx, embedding, storage.SearchOpts{
		Limit: 10,
		Repo:  "org/repo-a",
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	for _, r := range results {
		if r.Repo != "org/repo-a" {
			t.Errorf("expected repo 'org/repo-a', got %q", r.Repo)
		}
	}
}
