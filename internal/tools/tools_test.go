package tools_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MereWhiplash/engram-cogitator/internal/mcptypes"
	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/tools"
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
	// Simple filter simulation
	var results []types.Memory
	for _, mem := range m.memories {
		if !mem.IsValid {
			continue
		}
		if opts.Type != "" && mem.Type != opts.Type {
			continue
		}
		if opts.Area != "" && mem.Area != opts.Area {
			continue
		}
		results = append(results, mem)
		if len(results) >= opts.Limit {
			break
		}
	}
	return results, nil
}

func (m *mockStorage) List(ctx context.Context, opts types.ListOpts) ([]types.Memory, error) {
	var results []types.Memory
	for _, mem := range m.memories {
		if !opts.IncludeInvalid && !mem.IsValid {
			continue
		}
		if opts.Type != "" && mem.Type != opts.Type {
			continue
		}
		if opts.Area != "" && mem.Area != opts.Area {
			continue
		}
		results = append(results, mem)
		if opts.Limit > 0 && len(results) >= opts.Limit {
			break
		}
	}
	return results, nil
}

func (m *mockStorage) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
	for i := range m.memories {
		if m.memories[i].ID == id {
			m.memories[i].IsValid = false
			m.memories[i].SupersededBy = supersededBy
			return nil
		}
	}
	return types.ErrNotFound
}

func (m *mockStorage) Close() error {
	return nil
}

func newTestServer() (*mcp.Server, *mockStorage) {
	store := &mockStorage{}
	emb := &mockEmbedder{}
	svc := service.New(store, emb)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "0.0.1",
	}, nil)

	tools.Register(server, svc)
	return server, store
}

func TestAdd_Success(t *testing.T) {
	server, store := newTestServer()
	_ = server // server is used for registration

	// Create handler directly for testing
	svc := service.New(store, &mockEmbedder{})
	h := struct {
		svc *service.Service
	}{svc: svc}
	_ = h

	ctx := context.Background()

	// Test via service directly (tools.Handler is not exported)
	mem, err := svc.Add(ctx, types.TypeDecision, "auth", "Use JWT tokens", "Stateless auth")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if mem.Type != types.TypeDecision {
		t.Errorf("expected type 'decision', got %q", mem.Type)
	}
	if mem.Area != "auth" {
		t.Errorf("expected area 'auth', got %q", mem.Area)
	}
	if mem.Content != "Use JWT tokens" {
		t.Errorf("expected content 'Use JWT tokens', got %q", mem.Content)
	}
	if mem.Rationale != "Stateless auth" {
		t.Errorf("expected rationale 'Stateless auth', got %q", mem.Rationale)
	}
	if !mem.IsValid {
		t.Error("expected memory to be valid")
	}
}

func TestAdd_InvalidType(t *testing.T) {
	store := &mockStorage{}
	svc := service.New(store, &mockEmbedder{})

	ctx := context.Background()
	_, err := svc.Add(ctx, "invalid", "auth", "Should fail", "")
	if err == nil {
		t.Error("expected error for invalid type, got nil")
	}
}

func TestSearch_Success(t *testing.T) {
	store := &mockStorage{}
	svc := service.New(store, &mockEmbedder{})

	ctx := context.Background()

	// Add some memories
	_, _ = svc.Add(ctx, types.TypeDecision, "auth", "Use JWT", "")
	_, _ = svc.Add(ctx, types.TypeLearning, "auth", "JWT expiry gotcha", "")
	_, _ = svc.Add(ctx, types.TypePattern, "api", "REST conventions", "")

	// Search all
	results, err := svc.Search(ctx, "authentication", 10, "", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Search with type filter
	results, err = svc.Search(ctx, "auth", 10, types.TypeDecision, "")
	if err != nil {
		t.Fatalf("Search with type filter failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result with type filter, got %d", len(results))
	}

	// Search with area filter
	results, err = svc.Search(ctx, "patterns", 10, "", "api")
	if err != nil {
		t.Fatalf("Search with area filter failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result with area filter, got %d", len(results))
	}
}

func TestSearch_WithLimit(t *testing.T) {
	store := &mockStorage{}
	svc := service.New(store, &mockEmbedder{})

	ctx := context.Background()

	// Add multiple memories
	for i := 0; i < 10; i++ {
		_, _ = svc.Add(ctx, types.TypeDecision, "test", "Memory", "")
	}

	// Search with limit
	results, err := svc.Search(ctx, "test", 3, "", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results with limit, got %d", len(results))
	}
}

func TestList_Success(t *testing.T) {
	store := &mockStorage{}
	svc := service.New(store, &mockEmbedder{})

	ctx := context.Background()

	// Add some memories
	_, _ = svc.Add(ctx, types.TypeDecision, "auth", "Decision 1", "")
	_, _ = svc.Add(ctx, types.TypeLearning, "db", "Learning 1", "")

	// List all
	results, err := svc.List(ctx, 10, "", "", false)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// List with type filter
	results, err = svc.List(ctx, 10, types.TypeDecision, "", false)
	if err != nil {
		t.Fatalf("List with type filter failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result with type filter, got %d", len(results))
	}
}

func TestInvalidate_Success(t *testing.T) {
	store := &mockStorage{}
	svc := service.New(store, &mockEmbedder{})

	ctx := context.Background()

	// Add a memory
	mem, _ := svc.Add(ctx, types.TypeDecision, "auth", "Old decision", "")

	// Invalidate it
	err := svc.Invalidate(ctx, mem.ID, nil)
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	// List should not include invalidated by default
	results, _ := svc.List(ctx, 10, "", "", false)
	if len(results) != 0 {
		t.Errorf("expected 0 results after invalidate, got %d", len(results))
	}

	// List with includeInvalid should include it
	results, _ = svc.List(ctx, 10, "", "", true)
	if len(results) != 1 {
		t.Errorf("expected 1 result with includeInvalid, got %d", len(results))
	}
}

func TestInvalidate_WithSupersededBy(t *testing.T) {
	store := &mockStorage{}
	svc := service.New(store, &mockEmbedder{})

	ctx := context.Background()

	// Add old and new memories
	oldMem, _ := svc.Add(ctx, types.TypeDecision, "auth", "Use sessions", "")
	newMem, _ := svc.Add(ctx, types.TypeDecision, "auth", "Use JWT", "")

	// Invalidate old, superseded by new
	err := svc.Invalidate(ctx, oldMem.ID, &newMem.ID)
	if err != nil {
		t.Fatalf("Invalidate with supersededBy failed: %v", err)
	}

	// Verify supersededBy is set
	results, _ := svc.List(ctx, 10, "", "", true)
	for _, r := range results {
		if r.ID == oldMem.ID {
			if r.SupersededBy == nil || *r.SupersededBy != newMem.ID {
				t.Errorf("expected supersededBy to be %d, got %v", newMem.ID, r.SupersededBy)
			}
		}
	}
}

// TestMCPTypes verifies the shared types are correctly defined
func TestMCPTypes(t *testing.T) {
	// Verify AddInput fields
	input := mcptypes.AddInput{
		Type:      "decision",
		Area:      "auth",
		Content:   "Test content",
		Rationale: "Test rationale",
	}
	if input.Type != "decision" {
		t.Errorf("AddInput.Type mismatch")
	}

	// Verify SearchInput fields
	search := mcptypes.SearchInput{
		Query: "test",
		Limit: 5,
		Type:  "learning",
		Area:  "db",
	}
	if search.Query != "test" {
		t.Errorf("SearchInput.Query mismatch")
	}

	// Verify ListInput fields
	list := mcptypes.ListInput{
		Limit:          10,
		Type:           "pattern",
		Area:           "api",
		IncludeInvalid: true,
	}
	if !list.IncludeInvalid {
		t.Errorf("ListInput.IncludeInvalid mismatch")
	}

	// Verify InvalidateInput fields
	inv := mcptypes.InvalidateInput{
		ID:           1,
		SupersededBy: 2,
	}
	if inv.ID != 1 {
		t.Errorf("InvalidateInput.ID mismatch")
	}
}
