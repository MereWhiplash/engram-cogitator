// internal/api/handlers.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/MereWhiplash/engram-cogitator/internal/service"
)

// Handlers holds HTTP handler dependencies
type Handlers struct {
	svc *service.Service
}

// NewHandlers creates new API handlers
func NewHandlers(svc *service.Service) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handlers) respondError(w http.ResponseWriter, status int, msg string) {
	h.respondJSON(w, status, ErrorResponse{Error: msg})
}

// Health handles GET /health
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
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
		h.respondError(w, http.StatusInternalServerError, err.Error())
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
		h.respondError(w, http.StatusInternalServerError, err.Error())
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

	memType := r.URL.Query().Get("type")
	area := r.URL.Query().Get("area")
	repo := r.URL.Query().Get("repo")
	includeInvalid := r.URL.Query().Get("include_invalid") == "true"

	ctx := r.Context()

	memories, err := h.svc.ListWithRepo(ctx, limit, memType, area, repo, includeInvalid)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, ListResponse{Memories: memories})
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
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	msg := fmt.Sprintf("Memory %d has been invalidated.", id)
	if req.SupersededBy != nil {
		msg += fmt.Sprintf(" Superseded by memory %d.", *req.SupersededBy)
	}

	h.respondJSON(w, http.StatusOK, InvalidateResponse{Message: msg})
}
