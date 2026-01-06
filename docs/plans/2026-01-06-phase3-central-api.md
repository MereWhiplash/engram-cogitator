# Phase 3: Build Central API - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create an HTTP API server that exposes memory operations for team mode. The shim will call this API.

**Architecture:** Standalone HTTP server using standard library + chi router. Extracts git context from headers. Calls Service layer (same as MCP server uses).

**Tech Stack:** Go 1.23, chi router, existing Service/Storage/Embedder

**Prerequisites:** Phase 2 complete (storage backends working)

---

## Task 1: Add HTTP Router Dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add chi router**

Run:
```bash
go get github.com/go-chi/chi/v5
```

**Step 2: Tidy**

Run: `go mod tidy`

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add chi router dependency"
```

---

## Task 2: Create API Types

**Files:**
- Create: `internal/api/types.go`

**Step 1: Write request/response types**

```go
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
	Memories []storage.Memory `json:"memories"`
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
```

**Step 2: Verify it compiles**

Run: `go build ./internal/api/...`
Expected: Success

**Step 3: Commit**

```bash
git add internal/api/types.go
git commit -m "feat(api): add request/response types"
```

---

## Task 3: Create API Middleware

**Files:**
- Create: `internal/api/middleware.go`
- Create: `internal/api/middleware_test.go`

**Step 1: Write middleware**

```go
// internal/api/middleware.go
package api

import (
	"context"
	"net/http"
)

// Context keys for git info
type contextKey string

const (
	AuthorNameKey  contextKey = "author_name"
	AuthorEmailKey contextKey = "author_email"
	RepoKey        contextKey = "repo"
)

// GitContext extracts git info from headers and adds to context
func GitContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if name := r.Header.Get("X-EC-Author-Name"); name != "" {
			ctx = context.WithValue(ctx, AuthorNameKey, name)
		}
		if email := r.Header.Get("X-EC-Author-Email"); email != "" {
			ctx = context.WithValue(ctx, AuthorEmailKey, email)
		}
		if repo := r.Header.Get("X-EC-Repo"); repo != "" {
			ctx = context.WithValue(ctx, RepoKey, repo)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetAuthorName returns author name from context
func GetAuthorName(ctx context.Context) string {
	if v := ctx.Value(AuthorNameKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetAuthorEmail returns author email from context
func GetAuthorEmail(ctx context.Context) string {
	if v := ctx.Value(AuthorEmailKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetRepo returns repo from context
func GetRepo(ctx context.Context) string {
	if v := ctx.Value(RepoKey); v != nil {
		return v.(string)
	}
	return ""
}
```

**Step 2: Write test**

```go
// internal/api/middleware_test.go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/api"
)

func TestGitContext(t *testing.T) {
	var capturedName, capturedEmail, capturedRepo string

	handler := api.GitContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedName = api.GetAuthorName(r.Context())
		capturedEmail = api.GetAuthorEmail(r.Context())
		capturedRepo = api.GetRepo(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-EC-Author-Name", "Alice")
	req.Header.Set("X-EC-Author-Email", "alice@example.com")
	req.Header.Set("X-EC-Repo", "myorg/myrepo")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if capturedName != "Alice" {
		t.Errorf("expected name 'Alice', got %q", capturedName)
	}
	if capturedEmail != "alice@example.com" {
		t.Errorf("expected email 'alice@example.com', got %q", capturedEmail)
	}
	if capturedRepo != "myorg/myrepo" {
		t.Errorf("expected repo 'myorg/myrepo', got %q", capturedRepo)
	}
}
```

**Step 3: Run test**

Run: `go test ./internal/api/... -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/api/middleware.go internal/api/middleware_test.go
git commit -m "feat(api): add git context middleware"
```

---

## Task 4: Create API Handlers

**Files:**
- Create: `internal/api/handlers.go`
- Create: `internal/api/handlers_test.go`

**Step 1: Write handlers**

```go
// internal/api/handlers.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
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
```

**Step 2: Verify it compiles**

Run: `go build ./internal/api/...`
Expected: FAIL - service methods don't exist yet

**Step 3: Update Service to support team mode params**

```go
// Add to internal/service/service.go

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
func (s *Service) AddWithContext(ctx context.Context, params AddParams) (*storage.Memory, error) {
	textToEmbed := fmt.Sprintf("%s: %s", params.Area, params.Content)
	if params.Rationale != "" {
		textToEmbed += " " + params.Rationale
	}

	embedding, err := s.embedder.EmbedForStorage(textToEmbed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	mem := storage.Memory{
		Type:        params.Type,
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
func (s *Service) SearchWithRepo(ctx context.Context, query string, limit int, memType, area, repo string) ([]storage.Memory, error) {
	embedding, err := s.embedder.EmbedForSearch(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	opts := storage.SearchOpts{
		Limit: limit,
		Type:  memType,
		Area:  area,
		Repo:  repo,
	}

	return s.storage.Search(ctx, embedding, opts)
}

// ListWithRepo returns memories with optional repo filter
func (s *Service) ListWithRepo(ctx context.Context, limit int, memType, area, repo string, includeInvalid bool) ([]storage.Memory, error) {
	opts := storage.ListOpts{
		Limit:          limit,
		Type:           memType,
		Area:           area,
		Repo:           repo,
		IncludeInvalid: includeInvalid,
	}

	return s.storage.List(ctx, opts)
}
```

**Step 4: Verify it compiles**

Run: `go build ./internal/api/...`
Expected: Success

**Step 5: Write handler test**

```go
// internal/api/handlers_test.go
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/MereWhiplash/engram-cogitator/internal/api"
	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

type mockEmbedder struct{}

func (m *mockEmbedder) EmbedForStorage(text string) ([]float32, error) {
	return make([]float32, 768), nil
}

func (m *mockEmbedder) EmbedForSearch(query string) ([]float32, error) {
	return make([]float32, 768), nil
}

type mockStorage struct {
	memories []storage.Memory
	nextID   int64
}

func (m *mockStorage) Add(ctx context.Context, mem storage.Memory, embedding []float32) (*storage.Memory, error) {
	m.nextID++
	mem.ID = m.nextID
	mem.IsValid = true
	m.memories = append(m.memories, mem)
	return &mem, nil
}

func (m *mockStorage) Search(ctx context.Context, embedding []float32, opts storage.SearchOpts) ([]storage.Memory, error) {
	return m.memories, nil
}

func (m *mockStorage) List(ctx context.Context, opts storage.ListOpts) ([]storage.Memory, error) {
	return m.memories, nil
}

func (m *mockStorage) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func setupTestServer() (*api.Handlers, *chi.Mux) {
	store := &mockStorage{}
	emb := &mockEmbedder{}
	svc := service.New(store, emb)
	handlers := api.NewHandlers(svc)

	r := chi.NewRouter()
	r.Use(api.GitContext)
	r.Get("/health", handlers.Health)
	r.Post("/v1/memories", handlers.Add)
	r.Post("/v1/memories/search", handlers.Search)
	r.Get("/v1/memories", handlers.List)
	r.Put("/v1/memories/{id}/invalidate", handlers.Invalidate)

	return handlers, r
}

func TestHealth(t *testing.T) {
	_, r := setupTestServer()

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp api.HealthResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
}

func TestAdd(t *testing.T) {
	_, r := setupTestServer()

	body := api.AddRequest{
		Type:    "decision",
		Area:    "auth",
		Content: "Use JWT tokens",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/memories", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-EC-Author-Name", "Test User")
	req.Header.Set("X-EC-Author-Email", "test@example.com")
	req.Header.Set("X-EC-Repo", "testorg/testrepo")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp api.AddResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Memory == nil {
		t.Error("expected memory in response")
	}
	if resp.Memory.AuthorEmail != "test@example.com" {
		t.Errorf("expected author email 'test@example.com', got %q", resp.Memory.AuthorEmail)
	}
}

func TestSearch(t *testing.T) {
	_, r := setupTestServer()

	// First add a memory
	addBody := api.AddRequest{
		Type:    "decision",
		Area:    "auth",
		Content: "Use JWT tokens",
	}
	jsonBody, _ := json.Marshal(addBody)
	req := httptest.NewRequest("POST", "/v1/memories", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Now search
	searchBody := api.SearchRequest{
		Query: "authentication tokens",
		Limit: 5,
	}
	jsonBody, _ = json.Marshal(searchBody)
	req = httptest.NewRequest("POST", "/v1/memories/search", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp api.SearchResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Memories) != 1 {
		t.Errorf("expected 1 memory, got %d", len(resp.Memories))
	}
}
```

**Step 6: Run tests**

Run: `CGO_ENABLED=1 go test ./internal/api/... -v`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/api/handlers.go internal/api/handlers_test.go internal/service/service.go
git commit -m "feat(api): add HTTP handlers for memory operations"
```

---

## Task 5: Create API Server Entry Point

**Files:**
- Create: `cmd/api/main.go`

**Step 1: Write API server main**

```go
// cmd/api/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/MereWhiplash/engram-cogitator/internal/api"
	"github.com/MereWhiplash/engram-cogitator/internal/embedder"
	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

func main() {
	// Server flags
	addr := flag.String("addr", ":8080", "Server address")

	// Storage flags
	storageDriver := flag.String("storage-driver", "postgres", "Storage driver: postgres, mongodb")
	postgresDSN := flag.String("postgres-dsn", "", "PostgreSQL connection string")
	mongoURI := flag.String("mongodb-uri", "", "MongoDB connection URI")
	mongoDatabase := flag.String("mongodb-database", "engram", "MongoDB database name")

	// Embedder flags
	ollamaURL := flag.String("ollama-url", "http://localhost:11434", "Ollama API URL")
	embeddingModel := flag.String("embedding-model", "nomic-embed-text", "Ollama embedding model")

	// Migrate flag
	migrateOnly := flag.Bool("migrate", false, "Run migrations and exit")

	flag.Parse()

	ctx := context.Background()

	// Build storage config (no sqlite for API server - team mode only)
	cfg := storage.Config{
		Driver:          *storageDriver,
		PostgresDSN:     *postgresDSN,
		MongoDBURI:      *mongoURI,
		MongoDBDatabase: *mongoDatabase,
	}

	// Validate config
	if cfg.Driver == "sqlite" {
		log.Fatal("SQLite not supported for API server - use postgres or mongodb")
	}

	// Initialize storage
	store, err := storage.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// If migrate-only, exit now
	if *migrateOnly {
		log.Println("Migrations complete")
		return
	}

	// Initialize embedder
	emb := embedder.NewOllama(*ollamaURL, *embeddingModel)

	// Create service
	svc := service.New(store, emb)

	// Create handlers
	handlers := api.NewHandlers(svc)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(api.GitContext)

	// Routes
	r.Get("/health", handlers.Health)
	r.Route("/v1", func(r chi.Router) {
		r.Post("/memories", handlers.Add)
		r.Get("/memories", handlers.List)
		r.Post("/memories/search", handlers.Search)
		r.Put("/memories/{id}/invalidate", handlers.Invalidate)
	})

	// Create server
	srv := &http.Server{
		Addr:         *addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan bool)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}

		close(done)
	}()

	// Start server
	log.Printf("Starting API server on %s", *addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	<-done
	fmt.Println("Server stopped")
}
```

**Step 2: Verify build**

Run: `CGO_ENABLED=1 go build ./cmd/api`
Expected: Success

**Step 3: Commit**

```bash
git add cmd/api/main.go
git commit -m "feat(api): add central API server entry point"
```

---

## Task 6: Create API Dockerfile

**Files:**
- Create: `Dockerfile.api`

**Step 1: Write Dockerfile**

```dockerfile
# Dockerfile.api
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build API server
RUN CGO_ENABLED=0 GOOS=linux go build -o /ec-api ./cmd/api

# Runtime image
FROM gcr.io/distroless/static-debian12

COPY --from=builder /ec-api /ec-api

EXPOSE 8080

ENTRYPOINT ["/ec-api"]
```

**Step 2: Verify build**

Run: `docker build -f Dockerfile.api -t engram-api:local .`
Expected: Success

**Step 3: Commit**

```bash
git add Dockerfile.api
git commit -m "feat(api): add Dockerfile for API server"
```

---

## Summary

After Phase 3, you'll have:

```
cmd/
  server/main.go    # Solo mode MCP server
  api/main.go       # Team mode HTTP API server

internal/
  api/
    types.go        # Request/response types
    middleware.go   # Git context extraction
    handlers.go     # HTTP handlers

Dockerfile.api      # Container for API server
```

**API endpoints:**
```
GET  /health                    # K8s probes
POST /v1/memories               # Add memory
GET  /v1/memories               # List memories
POST /v1/memories/search        # Search memories
PUT  /v1/memories/:id/invalidate # Invalidate memory
```

**Headers (set by shim):**
```
X-EC-Author-Name: Alice Smith
X-EC-Author-Email: alice@example.com
X-EC-Repo: myorg/myrepo
```

**Next phase:** Build the MCP shim that extracts git info and forwards to this API.
