package storage

import (
	"context"

	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

// Storage defines the interface for memory persistence
type Storage interface {
	Add(ctx context.Context, mem types.Memory, embedding []float32) (*types.Memory, error)
	Search(ctx context.Context, embedding []float32, opts types.SearchOpts) ([]types.Memory, error)
	List(ctx context.Context, opts types.ListOpts) ([]types.Memory, error)
	Invalidate(ctx context.Context, id int64, supersededBy *int64) error
	Close() error
}
