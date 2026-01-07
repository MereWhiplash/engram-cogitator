package service

import (
	"context"
	"fmt"

	"github.com/MereWhiplash/engram-cogitator/internal/embedder"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
	"github.com/MereWhiplash/engram-cogitator/internal/types"
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
func (s *Service) Add(ctx context.Context, memType types.MemoryType, area, content, rationale string) (*types.Memory, error) {
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

	mem := types.Memory{
		Type:      memType,
		Area:      area,
		Content:   content,
		Rationale: rationale,
	}

	return s.storage.Add(ctx, mem, embedding)
}

// Search finds memories by semantic similarity
func (s *Service) Search(ctx context.Context, query string, limit int, memType types.MemoryType, area string) ([]types.Memory, error) {
	embedding, err := s.embedder.EmbedForSearch(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	opts := types.SearchOpts{
		Limit: limit,
		Type:  memType,
		Area:  area,
	}

	return s.storage.Search(ctx, embedding, opts)
}

// List returns recent memories
func (s *Service) List(ctx context.Context, limit int, memType types.MemoryType, area string, includeInvalid bool) ([]types.Memory, error) {
	opts := types.ListOpts{
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

// AddParams holds parameters for AddWithContext
type AddParams struct {
	Type        string
	Area        string
	Content     string
	Rationale   string
	AuthorName  string
	AuthorEmail string
	Repo        string
}

// AddWithContext creates a new memory with full context (for team mode)
func (s *Service) AddWithContext(ctx context.Context, params AddParams) (*types.Memory, error) {
	memType := types.MemoryType(params.Type)
	if err := memType.Validate(); err != nil {
		return nil, err
	}

	textToEmbed := fmt.Sprintf("%s: %s", params.Area, params.Content)
	if params.Rationale != "" {
		textToEmbed += " " + params.Rationale
	}

	embedding, err := s.embedder.EmbedForStorage(textToEmbed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	mem := types.Memory{
		Type:        memType,
		Area:        params.Area,
		Content:     params.Content,
		Rationale:   params.Rationale,
		AuthorName:  params.AuthorName,
		AuthorEmail: params.AuthorEmail,
		Repo:        params.Repo,
	}

	return s.storage.Add(ctx, mem, embedding)
}

// SearchWithRepo finds memories with optional repo filter
func (s *Service) SearchWithRepo(ctx context.Context, query string, limit int, memType, area, repo string) ([]types.Memory, error) {
	embedding, err := s.embedder.EmbedForSearch(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	opts := types.SearchOpts{
		Limit: limit,
		Type:  types.MemoryType(memType),
		Area:  area,
		Repo:  repo,
	}

	return s.storage.Search(ctx, embedding, opts)
}

// ListWithRepo returns memories with optional repo filter
func (s *Service) ListWithRepo(ctx context.Context, limit int, memType, area, repo string, includeInvalid bool, offset int) ([]types.Memory, error) {
	opts := types.ListOpts{
		Limit:          limit,
		Offset:         offset,
		Type:           types.MemoryType(memType),
		Area:           area,
		Repo:           repo,
		IncludeInvalid: includeInvalid,
	}

	return s.storage.List(ctx, opts)
}
