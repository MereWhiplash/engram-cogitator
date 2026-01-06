// internal/service/service.go
package service

import (
	"context"
	"fmt"

	"github.com/MereWhiplash/engram-cogitator/internal/embedder"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

// Service contains the business logic for memory operations
type Service struct {
	storage  storage.Storage
	embedder embedder.Embedder
}

// New creates a new Service
func New(store storage.Storage, emb embedder.Embedder) *Service {
	return &Service{
		storage:  store,
		embedder: emb,
	}
}

// Add creates a new memory entry
func (s *Service) Add(ctx context.Context, memType storage.MemoryType, area, content, rationale string) (*storage.Memory, error) {
	if err := memType.Validate(); err != nil {
		return nil, err
	}

	// Build text for embedding
	textToEmbed := fmt.Sprintf("%s: %s", area, content)
	if rationale != "" {
		textToEmbed += " " + rationale
	}

	embedding, err := s.embedder.EmbedForStorage(textToEmbed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	mem := storage.Memory{
		Type:      memType,
		Area:      area,
		Content:   content,
		Rationale: rationale,
	}

	return s.storage.Add(ctx, mem, embedding)
}

// Search finds memories by semantic similarity
func (s *Service) Search(ctx context.Context, query string, limit int, memType storage.MemoryType, area string) ([]storage.Memory, error) {
	embedding, err := s.embedder.EmbedForSearch(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	opts := storage.SearchOpts{
		Limit: limit,
		Type:  memType,
		Area:  area,
	}

	return s.storage.Search(ctx, embedding, opts)
}

// List returns recent memories
func (s *Service) List(ctx context.Context, limit int, memType storage.MemoryType, area string, includeInvalid bool) ([]storage.Memory, error) {
	opts := storage.ListOpts{
		Limit:          limit,
		Type:           memType,
		Area:           area,
		IncludeInvalid: includeInvalid,
	}

	return s.storage.List(ctx, opts)
}

// Invalidate marks a memory as invalid
func (s *Service) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
	return s.storage.Invalidate(ctx, id, supersededBy)
}

// Close cleans up resources
func (s *Service) Close() error {
	return s.storage.Close()
}
