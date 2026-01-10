package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MereWhiplash/engram-cogitator/internal/apitypes"
	"github.com/MereWhiplash/engram-cogitator/internal/client"
	"github.com/MereWhiplash/engram-cogitator/internal/gitinfo"
	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

func TestClient_Add_Success(t *testing.T) {
	expectedMem := &types.Memory{
		ID:        1,
		Type:      types.TypeDecision,
		Area:      "auth",
		Content:   "Use JWT",
		Rationale: "Stateless",
		IsValid:   true,
		CreatedAt: time.Now(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/memories" {
			t.Errorf("expected /v1/memories, got %s", r.URL.Path)
		}

		var req apitypes.AddRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Type != "decision" {
			t.Errorf("expected type 'decision', got %q", req.Type)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(apitypes.AddResponse{Memory: expectedMem})
	}))
	defer server.Close()

	c := client.New(server.URL, nil)
	mem, err := c.Add(context.Background(), "decision", "auth", "Use JWT", "Stateless")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if mem.ID != expectedMem.ID {
		t.Errorf("expected ID %d, got %d", expectedMem.ID, mem.ID)
	}
}

func TestClient_Add_WithGitInfo(t *testing.T) {
	var capturedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(apitypes.AddResponse{Memory: &types.Memory{ID: 1}})
	}))
	defer server.Close()

	gitInfo := &gitinfo.Info{
		AuthorName:  "Alice",
		AuthorEmail: "alice@example.com",
		Repo:        "myorg/myrepo",
	}

	c := client.New(server.URL, gitInfo)
	_, err := c.Add(context.Background(), "decision", "auth", "test", "")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if capturedHeaders.Get("X-EC-Author-Name") != "Alice" {
		t.Errorf("expected X-EC-Author-Name 'Alice', got %q", capturedHeaders.Get("X-EC-Author-Name"))
	}
	if capturedHeaders.Get("X-EC-Author-Email") != "alice@example.com" {
		t.Errorf("expected X-EC-Author-Email 'alice@example.com', got %q", capturedHeaders.Get("X-EC-Author-Email"))
	}
	if capturedHeaders.Get("X-EC-Repo") != "myorg/myrepo" {
		t.Errorf("expected X-EC-Repo 'myorg/myrepo', got %q", capturedHeaders.Get("X-EC-Repo"))
	}
}

func TestClient_Add_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apitypes.ErrorResponse{Error: "invalid type"})
	}))
	defer server.Close()

	c := client.New(server.URL, nil)
	_, err := c.Add(context.Background(), "invalid", "auth", "test", "")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Search_Success(t *testing.T) {
	expectedMems := []types.Memory{
		{ID: 1, Type: types.TypeDecision, Content: "Result 1"},
		{ID: 2, Type: types.TypeLearning, Content: "Result 2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/memories/search" {
			t.Errorf("expected /v1/memories/search, got %s", r.URL.Path)
		}

		var req apitypes.SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Query != "test query" {
			t.Errorf("expected query 'test query', got %q", req.Query)
		}
		if req.Limit != 5 {
			t.Errorf("expected limit 5, got %d", req.Limit)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(apitypes.SearchResponse{Memories: expectedMems})
	}))
	defer server.Close()

	c := client.New(server.URL, nil)
	mems, err := c.Search(context.Background(), "test query", 5, "", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(mems) != 2 {
		t.Errorf("expected 2 results, got %d", len(mems))
	}
}

func TestClient_Search_WithFilters(t *testing.T) {
	var capturedReq apitypes.SearchRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedReq)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(apitypes.SearchResponse{Memories: []types.Memory{}})
	}))
	defer server.Close()

	c := client.New(server.URL, nil)
	_, _ = c.Search(context.Background(), "test", 10, "decision", "auth")

	if capturedReq.Type != "decision" {
		t.Errorf("expected type 'decision', got %q", capturedReq.Type)
	}
	if capturedReq.Area != "auth" {
		t.Errorf("expected area 'auth', got %q", capturedReq.Area)
	}
}

func TestClient_List_Success(t *testing.T) {
	expectedMems := []types.Memory{
		{ID: 1, Type: types.TypeDecision, Content: "Memory 1"},
		{ID: 2, Type: types.TypeLearning, Content: "Memory 2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/memories" {
			t.Errorf("expected /v1/memories, got %s", r.URL.Path)
		}

		// Check query params
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %q", r.URL.Query().Get("limit"))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(apitypes.ListResponse{Memories: expectedMems})
	}))
	defer server.Close()

	c := client.New(server.URL, nil)
	mems, err := c.List(context.Background(), 10, "", "", false)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(mems) != 2 {
		t.Errorf("expected 2 results, got %d", len(mems))
	}
}

func TestClient_List_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("type") != "decision" {
			t.Errorf("expected type=decision, got %q", query.Get("type"))
		}
		if query.Get("area") != "auth" {
			t.Errorf("expected area=auth, got %q", query.Get("area"))
		}
		if query.Get("include_invalid") != "true" {
			t.Errorf("expected include_invalid=true, got %q", query.Get("include_invalid"))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(apitypes.ListResponse{Memories: []types.Memory{}})
	}))
	defer server.Close()

	c := client.New(server.URL, nil)
	_, _ = c.List(context.Background(), 5, "decision", "auth", true)
}

func TestClient_Invalidate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/v1/memories/123/invalidate" {
			t.Errorf("expected /v1/memories/123/invalidate, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(apitypes.InvalidateResponse{Message: "invalidated"})
	}))
	defer server.Close()

	c := client.New(server.URL, nil)
	err := c.Invalidate(context.Background(), 123, nil)
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}
}

func TestClient_Invalidate_WithSupersededBy(t *testing.T) {
	var capturedReq apitypes.InvalidateRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedReq)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(apitypes.InvalidateResponse{Message: "invalidated"})
	}))
	defer server.Close()

	c := client.New(server.URL, nil)
	supersededBy := int64(456)
	err := c.Invalidate(context.Background(), 123, &supersededBy)
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	if capturedReq.SupersededBy == nil || *capturedReq.SupersededBy != 456 {
		t.Errorf("expected superseded_by=456, got %v", capturedReq.SupersededBy)
	}
}

func TestClient_Invalidate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(apitypes.ErrorResponse{Error: "memory not found"})
	}))
	defer server.Close()

	c := client.New(server.URL, nil)
	err := c.Invalidate(context.Background(), 999, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_NetworkError(t *testing.T) {
	// Use a server that's already closed
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	c := client.New(server.URL, nil)

	_, err := c.Add(context.Background(), "decision", "auth", "test", "")
	if err == nil {
		t.Error("expected network error, got nil")
	}

	_, err = c.Search(context.Background(), "test", 5, "", "")
	if err == nil {
		t.Error("expected network error, got nil")
	}

	_, err = c.List(context.Background(), 10, "", "", false)
	if err == nil {
		t.Error("expected network error, got nil")
	}

	err = c.Invalidate(context.Background(), 1, nil)
	if err == nil {
		t.Error("expected network error, got nil")
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := client.New(server.URL, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := c.Add(ctx, "decision", "auth", "test", "")
	if err == nil {
		t.Error("expected context cancellation error, got nil")
	}
}
