package service_test

import (
	"context"
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

// mockEmbedder implements embedder.Embedder for testing
type mockEmbedder struct{}

func (m *mockEmbedder) EmbedForStorage(text string) ([]float32, error) {
	return make([]float32, 768), nil
}

func (m *mockEmbedder) EmbedForSearch(query string) ([]float32, error) {
	return make([]float32, 768), nil
}

// mockStorage implements storage.Storage for testing
type mockStorage struct {
	memories []types.Memory
	nextID   int64
}

func (m *mockStorage) Add(ctx context.Context, mem types.Memory, embedding []float32) (*types.Memory, error) {
	m.nextID++
	mem.ID = m.nextID
	mem.IsValid = true
	m.memories = append(m.memories, mem)
	return &mem, nil
}

func (m *mockStorage) Search(ctx context.Context, embedding []float32, opts types.SearchOpts) ([]types.Memory, error) {
	return m.memories, nil
}

func (m *mockStorage) List(ctx context.Context, opts types.ListOpts) ([]types.Memory, error) {
	return m.memories, nil
}

func (m *mockStorage) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestService_Add(t *testing.T) {
	store := &mockStorage{}
	emb := &mockEmbedder{}
	svc := service.New(store, emb)

	ctx := context.Background()
	mem, err := svc.Add(ctx, types.TypeDecision, "auth", "Use JWT", "Stateless")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if mem.Type != types.TypeDecision {
		t.Errorf("expected type 'decision', got %q", mem.Type)
	}
	if mem.Content != "Use JWT" {
		t.Errorf("expected content 'Use JWT', got %q", mem.Content)
	}
}

func TestService_Search(t *testing.T) {
	store := &mockStorage{}
	emb := &mockEmbedder{}
	svc := service.New(store, emb)

	ctx := context.Background()

	// Add a memory first
	_, err := svc.Add(ctx, types.TypeDecision, "auth", "Use JWT", "")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	results, err := svc.Search(ctx, "jwt tokens", 5, "", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestService_Add_InvalidType(t *testing.T) {
	store := &mockStorage{}
	emb := &mockEmbedder{}
	svc := service.New(store, emb)

	ctx := context.Background()
	_, err := svc.Add(ctx, "invalid", "auth", "Should fail", "")
	if err == nil {
		t.Error("expected error for invalid type, got nil")
	}
}
