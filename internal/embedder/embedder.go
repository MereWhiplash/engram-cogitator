// internal/embedder/embedder.go
package embedder

// Embedder generates vector embeddings for text
type Embedder interface {
	// EmbedForStorage creates an embedding optimized for document storage
	EmbedForStorage(text string) ([]float32, error)
	// EmbedForSearch creates an embedding optimized for search queries
	EmbedForSearch(query string) ([]float32, error)
}
