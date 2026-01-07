// internal/api/types.go
package api

import "github.com/MereWhiplash/engram-cogitator/internal/apitypes"

// Re-export types from internal/apitypes for backward compatibility.
// New code should import from internal/apitypes directly.
type (
	AddRequest         = apitypes.AddRequest
	AddResponse        = apitypes.AddResponse
	SearchRequest      = apitypes.SearchRequest
	SearchResponse     = apitypes.SearchResponse
	ListResponse       = apitypes.ListResponse
	PaginationInfo     = apitypes.PaginationInfo
	InvalidateRequest  = apitypes.InvalidateRequest
	InvalidateResponse = apitypes.InvalidateResponse
	ErrorResponse      = apitypes.ErrorResponse
	HealthResponse     = apitypes.HealthResponse
)
