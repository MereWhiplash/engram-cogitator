// internal/api/handlers_test.go
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/MereWhiplash/engram-cogitator/internal/api"
	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

type mockEmbedder struct{}

func (m *mockEmbedder) EmbedForStorage(text string) ([]float32, error) {
	return make([]float32, 768), nil
}

func (m *mockEmbedder) EmbedForSearch(query string) ([]float32, error) {
	return make([]float32, 768), nil
}

type mockStorage struct {
	memories []storage.Memory
	nextID   int64
}

func (m *mockStorage) Add(ctx context.Context, mem storage.Memory, embedding []float32) (*storage.Memory, error) {
	m.nextID++
	mem.ID = m.nextID
	mem.IsValid = true
	m.memories = append(m.memories, mem)
	return &mem, nil
}

func (m *mockStorage) Search(ctx context.Context, embedding []float32, opts storage.SearchOpts) ([]storage.Memory, error) {
	return m.memories, nil
}

func (m *mockStorage) List(ctx context.Context, opts storage.ListOpts) ([]storage.Memory, error) {
	return m.memories, nil
}

func (m *mockStorage) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
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

	var resp api.HealthResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
}

func TestAdd(t *testing.T) {
	_, r := setupTestServer()

	body := api.AddRequest{
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

	var resp api.AddResponse
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
	addBody := api.AddRequest{
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
	searchBody := api.SearchRequest{
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

	var resp api.SearchResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Memories) != 1 {
		t.Errorf("expected 1 memory, got %d", len(resp.Memories))
	}
}
