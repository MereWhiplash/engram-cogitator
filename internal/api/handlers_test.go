// internal/api/handlers_test.go
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/MereWhiplash/engram-cogitator/internal/api"
	"github.com/MereWhiplash/engram-cogitator/internal/apitypes"
	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

type mockEmbedder struct{}

func (m *mockEmbedder) EmbedForStorage(text string) ([]float32, error) {
	return make([]float32, 768), nil
}

func (m *mockEmbedder) EmbedForSearch(query string) ([]float32, error) {
	return make([]float32, 768), nil
}

type mockStorage struct {
	memories       []types.Memory
	nextID         int64
	invalidateErr  error
	invalidatedIDs []int64
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
	// Apply offset and limit
	start := opts.Offset
	if start > len(m.memories) {
		return []types.Memory{}, nil
	}
	end := start + opts.Limit
	if end > len(m.memories) {
		end = len(m.memories)
	}
	return m.memories[start:end], nil
}

func (m *mockStorage) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
	if m.invalidateErr != nil {
		return m.invalidateErr
	}
	// Check if memory exists
	found := false
	for _, mem := range m.memories {
		if mem.ID == id {
			found = true
			break
		}
	}
	if !found {
		return types.ErrNotFound
	}
	m.invalidatedIDs = append(m.invalidatedIDs, id)
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func setupTestServer() (*api.Handlers, *chi.Mux) {
	store := &mockStorage{}
	emb := &mockEmbedder{}
	svc := service.New(store, emb)
	handlers := api.NewHandlers(svc)

	r := chi.NewRouter()
	r.Use(api.GitContext)
	r.Get("/health", handlers.Health)
	r.Post("/v1/memories", handlers.Add)
	r.Post("/v1/memories/search", handlers.Search)
	r.Get("/v1/memories", handlers.List)
	r.Put("/v1/memories/{id}/invalidate", handlers.Invalidate)

	return handlers, r
}

func TestHealth(t *testing.T) {
	_, r := setupTestServer()

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp apitypes.HealthResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
}

func TestAdd(t *testing.T) {
	_, r := setupTestServer()

	body := apitypes.AddRequest{
		Type:    "decision",
		Area:    "auth",
		Content: "Use JWT tokens",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/memories", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-EC-Author-Name", "Test User")
	req.Header.Set("X-EC-Author-Email", "test@example.com")
	req.Header.Set("X-EC-Repo", "testorg/testrepo")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp apitypes.AddResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Memory == nil {
		t.Error("expected memory in response")
	}
	if resp.Memory.AuthorEmail != "test@example.com" {
		t.Errorf("expected author email 'test@example.com', got %q", resp.Memory.AuthorEmail)
	}
}

func TestSearch(t *testing.T) {
	_, r := setupTestServer()

	// First add a memory
	addBody := apitypes.AddRequest{
		Type:    "decision",
		Area:    "auth",
		Content: "Use JWT tokens",
	}
	jsonBody, _ := json.Marshal(addBody)
	req := httptest.NewRequest("POST", "/v1/memories", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Now search
	searchBody := apitypes.SearchRequest{
		Query: "authentication tokens",
		Limit: 5,
	}
	jsonBody, _ = json.Marshal(searchBody)
	req = httptest.NewRequest("POST", "/v1/memories/search", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp apitypes.SearchResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Memories) != 1 {
		t.Errorf("expected 1 memory, got %d", len(resp.Memories))
	}
}

func TestList(t *testing.T) {
	_, r := setupTestServer()

	// First add some memories
	for i := 0; i < 3; i++ {
		addBody := apitypes.AddRequest{
			Type:    "decision",
			Area:    "auth",
			Content: "Memory content",
		}
		jsonBody, _ := json.Marshal(addBody)
		req := httptest.NewRequest("POST", "/v1/memories", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}

	// Test basic list
	req := httptest.NewRequest("GET", "/v1/memories", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp apitypes.ListResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Memories) != 3 {
		t.Errorf("expected 3 memories, got %d", len(resp.Memories))
	}
	if resp.Pagination == nil {
		t.Error("expected pagination info")
	}

	// Test with limit
	req = httptest.NewRequest("GET", "/v1/memories?limit=2", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Memories) != 2 {
		t.Errorf("expected 2 memories with limit, got %d", len(resp.Memories))
	}
	if !resp.Pagination.HasMore {
		t.Error("expected HasMore to be true")
	}

	// Test with offset
	req = httptest.NewRequest("GET", "/v1/memories?limit=2&offset=2", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Memories) != 1 {
		t.Errorf("expected 1 memory with offset, got %d", len(resp.Memories))
	}
	if resp.Pagination.HasMore {
		t.Error("expected HasMore to be false")
	}
}

func TestInvalidate(t *testing.T) {
	_, r := setupTestServer()

	// First add a memory
	addBody := apitypes.AddRequest{
		Type:    "decision",
		Area:    "auth",
		Content: "Use JWT tokens",
	}
	jsonBody, _ := json.Marshal(addBody)
	req := httptest.NewRequest("POST", "/v1/memories", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var addResp apitypes.AddResponse
	json.NewDecoder(rr.Body).Decode(&addResp)
	memID := addResp.Memory.ID

	// Test successful invalidation
	req = httptest.NewRequest("PUT", fmt.Sprintf("/v1/memories/%d/invalidate", memID), nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var invResp apitypes.InvalidateResponse
	json.NewDecoder(rr.Body).Decode(&invResp)
	if invResp.Message == "" {
		t.Error("expected message in response")
	}

	// Test with superseded_by
	addBody2 := apitypes.AddRequest{
		Type:    "decision",
		Area:    "auth",
		Content: "Use OAuth2",
	}
	jsonBody, _ = json.Marshal(addBody2)
	req = httptest.NewRequest("POST", "/v1/memories", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var addResp2 apitypes.AddResponse
	json.NewDecoder(rr.Body).Decode(&addResp2)
	newMemID := addResp2.Memory.ID

	// Add another memory to invalidate with superseded_by
	addBody3 := apitypes.AddRequest{
		Type:    "decision",
		Area:    "auth",
		Content: "Third memory",
	}
	jsonBody, _ = json.Marshal(addBody3)
	req = httptest.NewRequest("POST", "/v1/memories", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var addResp3 apitypes.AddResponse
	json.NewDecoder(rr.Body).Decode(&addResp3)
	thirdMemID := addResp3.Memory.ID

	invBody := apitypes.InvalidateRequest{SupersededBy: &newMemID}
	jsonBody, _ = json.Marshal(invBody)
	req = httptest.NewRequest("PUT", fmt.Sprintf("/v1/memories/%d/invalidate", thirdMemID), bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 for superseded_by, got %d: %s", rr.Code, rr.Body.String())
	}

	// Test invalid ID format
	req = httptest.NewRequest("PUT", "/v1/memories/invalid/invalidate", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid ID, got %d", rr.Code)
	}

	// Test not found
	req = httptest.NewRequest("PUT", "/v1/memories/99999/invalidate", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for not found, got %d", rr.Code)
	}
}
