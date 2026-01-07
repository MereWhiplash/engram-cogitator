// internal/api/handlers.go
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

// Handlers holds HTTP handler dependencies
type Handlers struct {
	svc         *service.Service
	healthCheck func() error // optional health check function
}

// NewHandlers creates new API handlers
func NewHandlers(svc *service.Service) *Handlers {
	return &Handlers{svc: svc}
}

// SetHealthCheck sets an optional health check function for deeper health verification
func (h *Handlers) SetHealthCheck(check func() error) {
	h.healthCheck = check
}

func (h *Handlers) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func (h *Handlers) respondError(w http.ResponseWriter, status int, msg string) {
	h.respondJSON(w, status, ErrorResponse{Error: msg})
}

// logError logs an error with request context
func (h *Handlers) logError(r *http.Request, operation string, err error) {
	requestID := GetRequestID(r.Context())
	log.Printf("[%s] %s error: %v", requestID, operation, err)
}

// Health handles GET /health
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	if h.healthCheck != nil {
		if err := h.healthCheck(); err != nil {
			h.logError(r, "health", err)
			h.respondJSON(w, http.StatusServiceUnavailable, HealthResponse{Status: "unhealthy"})
			return
		}
	}
	h.respondJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

// Add handles POST /v1/memories
func (h *Handlers) Add(w http.ResponseWriter, r *http.Request) {
	var req AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Type == "" || req.Area == "" || req.Content == "" {
		h.respondError(w, http.StatusBadRequest, "type, area, and content are required")
		return
	}

	ctx := r.Context()

	// Create memory with git context
	mem, err := h.svc.AddWithContext(ctx, service.AddParams{
		Type:        req.Type,
		Area:        req.Area,
		Content:     req.Content,
		Rationale:   req.Rationale,
		AuthorName:  GetAuthorName(ctx),
		AuthorEmail: GetAuthorEmail(ctx),
		Repo:        GetRepo(ctx),
	})
	if err != nil {
		h.logError(r, "add", err)
		h.respondError(w, http.StatusInternalServerError, "failed to create memory")
		return
	}

	h.respondJSON(w, http.StatusCreated, AddResponse{Memory: mem})
}

// Search handles POST /v1/memories/search
func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Query == "" {
		h.respondError(w, http.StatusBadRequest, "query is required")
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 5
	}

	ctx := r.Context()

	memories, err := h.svc.SearchWithRepo(ctx, req.Query, limit, req.Type, req.Area, req.Repo)
	if err != nil {
		h.logError(r, "search", err)
		h.respondError(w, http.StatusInternalServerError, "failed to search memories")
		return
	}

	h.respondJSON(w, http.StatusOK, SearchResponse{Memories: memories})
}

// List handles GET /v1/memories
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	memType := r.URL.Query().Get("type")
	area := r.URL.Query().Get("area")
	repo := r.URL.Query().Get("repo")
	includeInvalid := r.URL.Query().Get("include_invalid") == "true"

	ctx := r.Context()

	// Request one extra to determine if there are more results
	memories, err := h.svc.ListWithRepo(ctx, limit+1, memType, area, repo, includeInvalid, offset)
	if err != nil {
		h.logError(r, "list", err)
		h.respondError(w, http.StatusInternalServerError, "failed to list memories")
		return
	}

	hasMore := len(memories) > limit
	if hasMore {
		memories = memories[:limit]
	}

	h.respondJSON(w, http.StatusOK, ListResponse{
		Memories: memories,
		Pagination: &PaginationInfo{
			Limit:   limit,
			Offset:  offset,
			HasMore: hasMore,
		},
	})
}

// Invalidate handles PUT /v1/memories/:id/invalidate
func (h *Handlers) Invalidate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid memory ID")
		return
	}

	var req InvalidateRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	ctx := r.Context()

	if err := h.svc.Invalidate(ctx, id, req.SupersededBy); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "memory not found")
			return
		}
		h.logError(r, "invalidate", err)
		h.respondError(w, http.StatusInternalServerError, "failed to invalidate memory")
		return
	}

	msg := fmt.Sprintf("Memory %d has been invalidated.", id)
	if req.SupersededBy != nil {
		msg += fmt.Sprintf(" Superseded by memory %d.", *req.SupersededBy)
	}

	h.respondJSON(w, http.StatusOK, InvalidateResponse{Message: msg})
}
