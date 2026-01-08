//go:build !cgo

package storage

import (
	"context"
	"fmt"

	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

// SQLite is a stub for non-CGO builds
type SQLite struct{}

var errNoCGO = fmt.Errorf("SQLite storage requires CGO (build with CGO_ENABLED=1)")

// NewSQLite returns an error in non-CGO builds
func NewSQLite(path string) (*SQLite, error) {
	return nil, errNoCGO
}

func (s *SQLite) Add(ctx context.Context, mem types.Memory, embedding []float32) (*types.Memory, error) {
	return nil, errNoCGO
}

func (s *SQLite) Search(ctx context.Context, embedding []float32, opts types.SearchOpts) ([]types.Memory, error) {
	return nil, errNoCGO
}

func (s *SQLite) List(ctx context.Context, opts types.ListOpts) ([]types.Memory, error) {
	return nil, errNoCGO
}

func (s *SQLite) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
	return errNoCGO
}

func (s *SQLite) Close() error {
	return nil
}
