package shim_test

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MereWhiplash/engram-cogitator/internal/mcptypes"
	"github.com/MereWhiplash/engram-cogitator/internal/shim"
	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

// mockAPIClient implements shim.APIClient for testing
type mockAPIClient struct {
	memories   []types.Memory
	nextID     int64
	addErr     error
	searchErr  error
	listErr    error
	invalidErr error
}

func (m *mockAPIClient) Add(ctx context.Context, memType, area, content, rationale string) (*types.Memory, error) {
	if m.addErr != nil {
		return nil, m.addErr
	}
	m.nextID++
	mem := types.Memory{
		ID:        m.nextID,
		Type:      types.MemoryType(memType),
		Area:      area,
		Content:   content,
		Rationale: rationale,
		IsValid:   true,
	}
	m.memories = append(m.memories, mem)
	return &mem, nil
}

func (m *mockAPIClient) Search(ctx context.Context, query string, limit int, memType, area string) ([]types.Memory, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	var results []types.Memory
	for _, mem := range m.memories {
		if memType != "" && string(mem.Type) != memType {
			continue
		}
		if area != "" && mem.Area != area {
			continue
		}
		results = append(results, mem)
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (m *mockAPIClient) List(ctx context.Context, limit int, memType, area string, includeInvalid bool) ([]types.Memory, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var results []types.Memory
	for _, mem := range m.memories {
		if !includeInvalid && !mem.IsValid {
			continue
		}
		if memType != "" && string(mem.Type) != memType {
			continue
		}
		if area != "" && mem.Area != area {
			continue
		}
		results = append(results, mem)
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (m *mockAPIClient) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
	if m.invalidErr != nil {
		return m.invalidErr
	}
	for i := range m.memories {
		if m.memories[i].ID == id {
			m.memories[i].IsValid = false
			m.memories[i].SupersededBy = supersededBy
			return nil
		}
	}
	return types.ErrNotFound
}

func TestShimHandler_Add_Success(t *testing.T) {
	client := &mockAPIClient{}
	handler := shim.NewHandler(client)

	ctx := context.Background()
	input := mcptypes.AddInput{
		Type:      "decision",
		Area:      "auth",
		Content:   "Use JWT tokens",
		Rationale: "Stateless",
	}

	result, output, err := handler.Add(ctx, nil, input)
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("Add returned error result: %v", result.Content)
	}
	if output.Memory == nil {
		t.Fatal("Add returned nil memory")
	}
	if output.Memory.Type != types.TypeDecision {
		t.Errorf("expected type 'decision', got %q", output.Memory.Type)
	}
}

func TestShimHandler_Add_MissingFields(t *testing.T) {
	client := &mockAPIClient{}
	handler := shim.NewHandler(client)

	ctx := context.Background()

	tests := []struct {
		name  string
		input mcptypes.AddInput
	}{
		{"missing type", mcptypes.AddInput{Area: "auth", Content: "test"}},
		{"missing area", mcptypes.AddInput{Type: "decision", Content: "test"}},
		{"missing content", mcptypes.AddInput{Type: "decision", Area: "auth"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := handler.Add(ctx, nil, tt.input)
			if !result.IsError {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

func TestShimHandler_Add_ClientError(t *testing.T) {
	client := &mockAPIClient{addErr: errors.New("connection failed")}
	handler := shim.NewHandler(client)

	ctx := context.Background()
	input := mcptypes.AddInput{
		Type:    "decision",
		Area:    "auth",
		Content: "Test",
	}

	result, _, _ := handler.Add(ctx, nil, input)
	if !result.IsError {
		t.Error("expected error result when client fails")
	}
}

func TestShimHandler_Search_Success(t *testing.T) {
	client := &mockAPIClient{}
	// Pre-populate with memories
	client.Add(context.Background(), "decision", "auth", "Use JWT", "")
	client.Add(context.Background(), "learning", "auth", "JWT gotcha", "")

	handler := shim.NewHandler(client)

	ctx := context.Background()
	input := mcptypes.SearchInput{
		Query: "authentication",
		Limit: 10,
	}

	result, output, err := handler.Search(ctx, nil, input)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("Search returned error result: %v", result.Content)
	}
	if len(output.Memories) != 2 {
		t.Errorf("expected 2 memories, got %d", len(output.Memories))
	}
}

func TestShimHandler_Search_MissingQuery(t *testing.T) {
	client := &mockAPIClient{}
	handler := shim.NewHandler(client)

	ctx := context.Background()
	input := mcptypes.SearchInput{Limit: 5}

	result, _, _ := handler.Search(ctx, nil, input)
	if !result.IsError {
		t.Error("expected error for missing query")
	}
}

func TestShimHandler_Search_DefaultLimit(t *testing.T) {
	client := &mockAPIClient{}
	// Add more than default limit
	for i := 0; i < 10; i++ {
		client.Add(context.Background(), "decision", "test", "Memory", "")
	}

	handler := shim.NewHandler(client)

	ctx := context.Background()
	input := mcptypes.SearchInput{Query: "test"} // No limit specified

	_, output, _ := handler.Search(ctx, nil, input)
	if len(output.Memories) != 5 { // Default limit is 5
		t.Errorf("expected 5 memories (default limit), got %d", len(output.Memories))
	}
}

func TestShimHandler_Search_NoResults(t *testing.T) {
	client := &mockAPIClient{}
	handler := shim.NewHandler(client)

	ctx := context.Background()
	input := mcptypes.SearchInput{Query: "nonexistent"}

	result, output, _ := handler.Search(ctx, nil, input)
	if result.IsError {
		t.Error("empty results should not be an error")
	}
	if len(output.Memories) != 0 {
		t.Errorf("expected 0 memories, got %d", len(output.Memories))
	}
}

func TestShimHandler_List_Success(t *testing.T) {
	client := &mockAPIClient{}
	client.Add(context.Background(), "decision", "auth", "Decision 1", "")
	client.Add(context.Background(), "learning", "db", "Learning 1", "")

	handler := shim.NewHandler(client)

	ctx := context.Background()
	input := mcptypes.ListInput{Limit: 10}

	result, output, err := handler.List(ctx, nil, input)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("List returned error result: %v", result.Content)
	}
	if len(output.Memories) != 2 {
		t.Errorf("expected 2 memories, got %d", len(output.Memories))
	}
}

func TestShimHandler_List_DefaultLimit(t *testing.T) {
	client := &mockAPIClient{}
	for i := 0; i < 15; i++ {
		client.Add(context.Background(), "decision", "test", "Memory", "")
	}

	handler := shim.NewHandler(client)

	ctx := context.Background()
	input := mcptypes.ListInput{} // No limit specified

	_, output, _ := handler.List(ctx, nil, input)
	if len(output.Memories) != 10 { // Default limit is 10
		t.Errorf("expected 10 memories (default limit), got %d", len(output.Memories))
	}
}

func TestShimHandler_Invalidate_Success(t *testing.T) {
	client := &mockAPIClient{}
	mem, _ := client.Add(context.Background(), "decision", "auth", "Old", "")

	handler := shim.NewHandler(client)

	ctx := context.Background()
	input := mcptypes.InvalidateInput{ID: mem.ID}

	result, output, err := handler.Invalidate(ctx, nil, input)
	if err != nil {
		t.Fatalf("Invalidate returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("Invalidate returned error result: %v", result.Content)
	}
	if output.Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestShimHandler_Invalidate_MissingID(t *testing.T) {
	client := &mockAPIClient{}
	handler := shim.NewHandler(client)

	ctx := context.Background()
	input := mcptypes.InvalidateInput{} // ID is 0

	result, _, _ := handler.Invalidate(ctx, nil, input)
	if !result.IsError {
		t.Error("expected error for missing ID")
	}
}

func TestShimHandler_Invalidate_WithSupersededBy(t *testing.T) {
	client := &mockAPIClient{}
	oldMem, _ := client.Add(context.Background(), "decision", "auth", "Old", "")
	newMem, _ := client.Add(context.Background(), "decision", "auth", "New", "")

	handler := shim.NewHandler(client)

	ctx := context.Background()
	input := mcptypes.InvalidateInput{
		ID:           oldMem.ID,
		SupersededBy: newMem.ID,
	}

	result, output, _ := handler.Invalidate(ctx, nil, input)
	if result.IsError {
		t.Fatalf("Invalidate returned error result: %v", result.Content)
	}
	// Message should mention supersededBy
	if output.Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestShimRegister(t *testing.T) {
	client := &mockAPIClient{}
	handler := shim.NewHandler(client)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "0.0.1",
	}, nil)

	// Should not panic
	shim.Register(server, handler)
}
