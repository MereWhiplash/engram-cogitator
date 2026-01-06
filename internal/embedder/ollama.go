// internal/embedder/ollama.go
package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Ollama implements Embedder using Ollama API
type Ollama struct {
	baseURL string
	model   string
	http    *http.Client
}

type embeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// NewOllama creates a new Ollama embedder
func NewOllama(baseURL, model string) *Ollama {
	return &Ollama{
		baseURL: baseURL,
		model:   model,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (o *Ollama) embed(text string) ([]float32, error) {
	reqBody := embeddingRequest{
		Model:  o.model,
		Prompt: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := o.http.Post(
		fmt.Sprintf("%s/api/embeddings", o.baseURL),
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

	var embResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return embResp.Embedding, nil
}

func (o *Ollama) EmbedForStorage(text string) ([]float32, error) {
	if o.model == "nomic-embed-text" {
		return o.embed("search_document: " + text)
	}
	return o.embed(text)
}

func (o *Ollama) EmbedForSearch(query string) ([]float32, error) {
	if o.model == "nomic-embed-text" {
		return o.embed("search_query: " + query)
	}
	return o.embed(query)
}
