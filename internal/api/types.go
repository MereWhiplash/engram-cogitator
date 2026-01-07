// internal/api/types.go
package api

import "github.com/MereWhiplash/engram-cogitator/internal/storage"

// AddRequest is the request body for POST /v1/memories
type AddRequest struct {
	Type      string `json:"type"`
	Area      string `json:"area"`
	Content   string `json:"content"`
	Rationale string `json:"rationale,omitempty"`
}

// AddResponse is the response for POST /v1/memories
type AddResponse struct {
	Memory *storage.Memory `json:"memory"`
}

// SearchRequest is the request body for POST /v1/memories/search
type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
	Type  string `json:"type,omitempty"`
	Area  string `json:"area,omitempty"`
	Repo  string `json:"repo,omitempty"` // empty = all repos
}

// SearchResponse is the response for POST /v1/memories/search
type SearchResponse struct {
	Memories []storage.Memory `json:"memories"`
}

// ListResponse is the response for GET /v1/memories
type ListResponse struct {
	Memories   []storage.Memory `json:"memories"`
	Pagination *PaginationInfo  `json:"pagination,omitempty"`
}

// PaginationInfo provides pagination metadata
type PaginationInfo struct {
	Limit  int  `json:"limit"`
	Offset int  `json:"offset"`
	HasMore bool `json:"has_more"`
}

// InvalidateRequest is the request body for PUT /v1/memories/:id/invalidate
type InvalidateRequest struct {
	SupersededBy *int64 `json:"superseded_by,omitempty"`
}

// InvalidateResponse is the response for PUT /v1/memories/:id/invalidate
type InvalidateResponse struct {
	Message string `json:"message"`
}

// ErrorResponse is returned on errors
type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthResponse is the response for GET /health
type HealthResponse struct {
	Status string `json:"status"`
}
