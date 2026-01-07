package embedder

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllama_EmbedForStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := embeddingResponse{
			Embedding: make([]float32, 768),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOllama(server.URL, "nomic-embed-text")
	emb, err := client.EmbedForStorage("test content")
	if err != nil {
		t.Fatalf("EmbedForStorage failed: %v", err)
	}

	if len(emb) != 768 {
		t.Errorf("expected 768 dimensions, got %d", len(emb))
	}
}

func TestOllama_EmbedForSearch(t *testing.T) {
	var receivedPrompt string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req embeddingRequest
		json.NewDecoder(r.Body).Decode(&req)
		receivedPrompt = req.Prompt

		resp := embeddingResponse{
			Embedding: make([]float32, 768),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOllama(server.URL, "nomic-embed-text")
	_, err := client.EmbedForSearch("test query")
	if err != nil {
		t.Fatalf("EmbedForSearch failed: %v", err)
	}

	expected := "search_query: test query"
	if receivedPrompt != expected {
		t.Errorf("expected prompt %q, got %q", expected, receivedPrompt)
	}
}

func TestOllama_EmbedForStorage_Prefix(t *testing.T) {
	var receivedPrompt string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req embeddingRequest
		json.NewDecoder(r.Body).Decode(&req)
		receivedPrompt = req.Prompt

		resp := embeddingResponse{
			Embedding: make([]float32, 768),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOllama(server.URL, "nomic-embed-text")
	_, err := client.EmbedForStorage("test document")
	if err != nil {
		t.Fatalf("EmbedForStorage failed: %v", err)
	}

	expected := "search_document: test document"
	if receivedPrompt != expected {
		t.Errorf("expected prompt %q, got %q", expected, receivedPrompt)
	}
}

func TestOllama_OllamaDown(t *testing.T) {
	client := NewOllama("http://localhost:99999", "nomic-embed-text")
	_, err := client.EmbedForStorage("test content")

	if err == nil {
		t.Error("expected error when Ollama is down")
	}
}

func TestOllama_NonNomicModel(t *testing.T) {
	var receivedPrompt string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req embeddingRequest
		json.NewDecoder(r.Body).Decode(&req)
		receivedPrompt = req.Prompt

		resp := embeddingResponse{
			Embedding: make([]float32, 768),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Non-nomic model should not add prefix
	client := NewOllama(server.URL, "other-model")
	_, err := client.EmbedForStorage("test content")
	if err != nil {
		t.Fatalf("EmbedForStorage failed: %v", err)
	}

	// Should NOT have prefix for non-nomic models
	if receivedPrompt != "test content" {
		t.Errorf("expected prompt %q without prefix, got %q", "test content", receivedPrompt)
	}
}

func TestOllama_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewOllama(server.URL, "nomic-embed-text")
	_, err := client.EmbedForStorage("test content")

	if err == nil {
		t.Error("expected error on HTTP 500")
	}
}
