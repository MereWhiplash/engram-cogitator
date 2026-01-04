package embed

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbedForStorage(t *testing.T) {
	// Mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Return mock 768-dim embedding
		resp := EmbeddingResponse{
			Embedding: make([]float32, 768),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, "nomic-embed-text")
	emb, err := client.EmbedForStorage("test content")
	if err != nil {
		t.Fatalf("EmbedForStorage failed: %v", err)
	}

	if len(emb) != 768 {
		t.Errorf("expected 768 dimensions, got %d", len(emb))
	}
}

func TestEmbedDimensionMismatch(t *testing.T) {
	// Mock server returning wrong dimensions
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := EmbeddingResponse{
			Embedding: make([]float32, 384), // Wrong!
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, "nomic-embed-text")
	_, err := client.EmbedForStorage("test content")

	// Should succeed (we don't validate dimensions in embed layer)
	// The DB layer will catch dimension mismatch
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmbedOllamaDown(t *testing.T) {
	client := New("http://localhost:99999", "nomic-embed-text")
	_, err := client.EmbedForStorage("test content")

	if err == nil {
		t.Error("expected error when Ollama is down")
	}
}

func TestEmbedForSearch(t *testing.T) {
	var receivedPrompt string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req EmbeddingRequest
		json.NewDecoder(r.Body).Decode(&req)
		receivedPrompt = req.Prompt

		resp := EmbeddingResponse{
			Embedding: make([]float32, 768),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL, "nomic-embed-text")
	_, err := client.EmbedForSearch("test query")
	if err != nil {
		t.Fatalf("EmbedForSearch failed: %v", err)
	}

	// nomic-embed-text should prepend "search_query: "
	expected := "search_query: test query"
	if receivedPrompt != expected {
		t.Errorf("expected prompt %q, got %q", expected, receivedPrompt)
	}
}
