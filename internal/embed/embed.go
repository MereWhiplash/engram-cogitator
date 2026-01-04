package embed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with Ollama for embeddings
type Client struct {
	baseURL string
	model   string
	http    *http.Client
}

// EmbeddingRequest represents the Ollama embedding API request
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse represents the Ollama embedding API response
type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// New creates a new Ollama embedding client
func New(baseURL, model string) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Embed generates an embedding for the given text
func (c *Client) Embed(text string) ([]float32, error) {
	reqBody := EmbeddingRequest{
		Model:  c.model,
		Prompt: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.http.Post(
		fmt.Sprintf("%s/api/embeddings", c.baseURL),
		"application/json",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var embResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return embResp.Embedding, nil
}

// EmbedForSearch creates an embedding optimized for search queries
// For nomic-embed-text, prepend "search_query: " for queries
func (c *Client) EmbedForSearch(query string) ([]float32, error) {
	// nomic-embed-text uses prefixes for better search performance
	if c.model == "nomic-embed-text" {
		return c.Embed("search_query: " + query)
	}
	return c.Embed(query)
}

// EmbedForStorage creates an embedding optimized for document storage
// For nomic-embed-text, prepend "search_document: " for documents
func (c *Client) EmbedForStorage(document string) ([]float32, error) {
	// nomic-embed-text uses prefixes for better search performance
	if c.model == "nomic-embed-text" {
		return c.Embed("search_document: " + document)
	}
	return c.Embed(document)
}
