// internal/types/types.go
// Package types contains shared data types that have no CGO dependencies.
// This allows packages like the shim to use Memory type without pulling in sqlite-vec.
package types

import (
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned when a memory is not found
var ErrNotFound = errors.New("memory not found")

// MemoryType represents the type of memory entry
type MemoryType string

const (
	TypeDecision MemoryType = "decision"
	TypeLearning MemoryType = "learning"
	TypePattern  MemoryType = "pattern"
)

// Valid returns true if the MemoryType is a known valid type
func (t MemoryType) Valid() bool {
	switch t {
	case TypeDecision, TypeLearning, TypePattern:
		return true
	}
	return false
}

// Validate returns an error if the MemoryType is invalid
func (t MemoryType) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("invalid memory type %q: must be decision, learning, or pattern", t)
	}
	return nil
}

// Memory represents a stored memory entry
type Memory struct {
	ID           int64      `json:"id"`
	Type         MemoryType `json:"type"`
	Area         string     `json:"area"`
	Content      string     `json:"content"`
	Rationale    string     `json:"rationale,omitempty"`
	IsValid      bool       `json:"is_valid"`
	SupersededBy *int64     `json:"superseded_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	// Team mode fields (optional, empty for solo mode)
	AuthorName  string `json:"author_name,omitempty"`
	AuthorEmail string `json:"author_email,omitempty"`
	Repo        string `json:"repo,omitempty"`
}

// SearchOpts configures search behavior
type SearchOpts struct {
	Limit int
	Type  MemoryType
	Area  string
	Repo  string // team mode only
}

// ListOpts configures list behavior
type ListOpts struct {
	Limit          int
	Offset         int
	Type           MemoryType
	Area           string
	Repo           string // team mode only
	IncludeInvalid bool
}
