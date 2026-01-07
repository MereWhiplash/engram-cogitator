package storage

import (
	"context"

	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

// Re-export types from internal/types for backward compatibility.
// New code should import from internal/types directly.
type (
	MemoryType = types.MemoryType
	Memory     = types.Memory
	SearchOpts = types.SearchOpts
	ListOpts   = types.ListOpts
)

const (
	TypeDecision = types.TypeDecision
	TypeLearning = types.TypeLearning
	TypePattern  = types.TypePattern
)

var ErrNotFound = types.ErrNotFound

// Storage defines the interface for memory persistence
type Storage interface {
	Add(ctx context.Context, mem Memory, embedding []float32) (*Memory, error)
	Search(ctx context.Context, embedding []float32, opts SearchOpts) ([]Memory, error)
	List(ctx context.Context, opts ListOpts) ([]Memory, error)
	Invalidate(ctx context.Context, id int64, supersededBy *int64) error
	Close() error
}
