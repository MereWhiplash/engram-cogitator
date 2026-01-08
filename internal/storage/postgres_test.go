package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MereWhiplash/engram-cogitator/internal/storage"
	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

// cleanupPostgres removes all test data before each test
func cleanupPostgres(t *testing.T, dsn string) {
	t.Helper()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect for cleanup: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, "DELETE FROM memory_embeddings")
	if err != nil {
		t.Fatalf("failed to cleanup embeddings: %v", err)
	}
	_, err = pool.Exec(ctx, "DELETE FROM memories")
	if err != nil {
		t.Fatalf("failed to cleanup memories: %v", err)
	}
}

func TestPostgresStorage_Add(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set, skipping Postgres tests")
	}
	cleanupPostgres(t, dsn)

	ctx := context.Background()
	store, err := storage.NewPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	mem := types.Memory{
		Type:        types.TypeDecision,
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
	cleanupPostgres(t, dsn)

	ctx := context.Background()
	store, err := storage.NewPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	// Add a memory
	mem := types.Memory{
		Type:        types.TypeDecision,
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
	results, err := store.Search(ctx, embedding, types.SearchOpts{Limit: 5})
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
	cleanupPostgres(t, dsn)

	ctx := context.Background()
	store, err := storage.NewPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	embedding := make([]float32, 768)

	// Add memories to different repos
	mem1 := types.Memory{
		Type:        types.TypeDecision,
		Area:        "auth",
		Content:     "Repo A decision",
		AuthorName:  "User A",
		AuthorEmail: "a@example.com",
		Repo:        "org/repo-a",
	}
	mem2 := types.Memory{
		Type:        types.TypeDecision,
		Area:        "auth",
		Content:     "Repo B decision",
		AuthorName:  "User B",
		AuthorEmail: "b@example.com",
		Repo:        "org/repo-b",
	}

	_, _ = store.Add(ctx, mem1, embedding)
	_, _ = store.Add(ctx, mem2, embedding)

	// Search scoped to repo-a
	results, err := store.Search(ctx, embedding, types.SearchOpts{
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
