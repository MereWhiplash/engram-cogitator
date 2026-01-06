# Phase 1: Refactor Internals - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extract clean interfaces for Storage and Embedder, add Service layer, without changing external behavior.

**Architecture:** Current code has db/embed packages directly used by tool handlers. We extract interfaces so team mode can swap implementations (Postgres, MongoDB) without changing business logic.

**Tech Stack:** Go 1.23, existing dependencies only (no new deps in Phase 1)

---

## Task 1: Create Storage Interface

**Files:**
- Create: `internal/storage/storage.go`
- Test: `internal/storage/storage_test.go`

**Step 1: Write the interface and types**

```go
// internal/storage/storage.go
package storage

import (
	"context"
	"time"
)

// Memory represents a stored memory entry
type Memory struct {
	ID           int64     `json:"id"`
	Type         string    `json:"type"`
	Area         string    `json:"area"`
	Content      string    `json:"content"`
	Rationale    string    `json:"rationale,omitempty"`
	IsValid      bool      `json:"is_valid"`
	SupersededBy *int64    `json:"superseded_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	// Team mode fields (optional, empty for solo mode)
	AuthorName  string `json:"author_name,omitempty"`
	AuthorEmail string `json:"author_email,omitempty"`
	Repo        string `json:"repo,omitempty"`
}

// SearchOpts configures search behavior
type SearchOpts struct {
	Limit int
	Type  string
	Area  string
	Repo  string // team mode only
}

// ListOpts configures list behavior
type ListOpts struct {
	Limit          int
	Type           string
	Area           string
	Repo           string // team mode only
	IncludeInvalid bool
}

// Storage defines the interface for memory persistence
type Storage interface {
	Add(ctx context.Context, mem Memory, embedding []float32) (*Memory, error)
	Search(ctx context.Context, embedding []float32, opts SearchOpts) ([]Memory, error)
	List(ctx context.Context, opts ListOpts) ([]Memory, error)
	Invalidate(ctx context.Context, id int64, supersededBy *int64) error
	Close() error
}
```

**Step 2: Verify it compiles**

Run: `CGO_ENABLED=1 go build ./internal/storage/...`
Expected: Success (no output)

**Step 3: Commit**

```bash
git add internal/storage/storage.go
git commit -m "feat(storage): add Storage interface and types"
```

---

## Task 2: Create SQLite Implementation of Storage

**Files:**
- Create: `internal/storage/sqlite.go`
- Create: `internal/storage/sqlite_test.go`

**Step 1: Write failing test**

```go
// internal/storage/sqlite_test.go
package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

func TestSQLiteStorage_Add(t *testing.T) {
	// Use temp file for test database
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	store, err := storage.NewSQLite(f.Name())
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	mem := storage.Memory{
		Type:      "decision",
		Area:      "auth",
		Content:   "Use JWT tokens",
		Rationale: "Stateless auth",
	}
	embedding := make([]float32, 768)

	result, err := store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if result.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if result.Type != "decision" {
		t.Errorf("expected type 'decision', got %q", result.Type)
	}
	if !result.IsValid {
		t.Error("expected IsValid to be true")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=1 go test ./internal/storage/... -v`
Expected: FAIL with "undefined: storage.NewSQLite"

**Step 3: Write SQLite implementation**

```go
// internal/storage/sqlite.go
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

// SQLite implements Storage using SQLite with sqlite-vec
type SQLite struct {
	conn *sql.DB
}

// NewSQLite creates a new SQLite storage
func NewSQLite(path string) (*SQLite, error) {
	sqlite_vec.Auto()

	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	s := &SQLite{conn: conn}
	if err := s.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return s, nil
}

func (s *SQLite) initSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS memories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL CHECK(type IN ('decision', 'learning', 'pattern')),
			area TEXT NOT NULL,
			content TEXT NOT NULL,
			rationale TEXT,
			is_valid BOOLEAN NOT NULL DEFAULT TRUE,
			superseded_by INTEGER REFERENCES memories(id),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
		CREATE INDEX IF NOT EXISTS idx_memories_area ON memories(area);
		CREATE INDEX IF NOT EXISTS idx_memories_is_valid ON memories(is_valid);

		CREATE VIRTUAL TABLE IF NOT EXISTS memory_embeddings USING vec0(
			memory_id INTEGER PRIMARY KEY,
			embedding FLOAT[768]
		);
	`
	_, err := s.conn.Exec(schema)
	return err
}

func (s *SQLite) Close() error {
	return s.conn.Close()
}

func (s *SQLite) Add(ctx context.Context, mem Memory, embedding []float32) (*Memory, error) {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO memories (type, area, content, rationale) VALUES (?, ?, ?, ?)`,
		mem.Type, mem.Area, mem.Content, mem.Rationale,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert memory: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO memory_embeddings (memory_id, embedding) VALUES (?, ?)`,
		id, string(embeddingJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert embedding: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &Memory{
		ID:        id,
		Type:      mem.Type,
		Area:      mem.Area,
		Content:   mem.Content,
		Rationale: mem.Rationale,
		IsValid:   true,
		CreatedAt: time.Now(),
	}, nil
}

func (s *SQLite) Search(ctx context.Context, embedding []float32, opts SearchOpts) ([]Memory, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 5
	}

	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding: %w", err)
	}

	query := `
		SELECT m.id, m.type, m.area, m.content, m.rationale, m.is_valid, m.superseded_by, m.created_at
		FROM memories m
		JOIN memory_embeddings e ON m.id = e.memory_id
		WHERE m.is_valid = TRUE
	`
	args := []interface{}{}

	if opts.Type != "" {
		query += " AND m.type = ?"
		args = append(args, opts.Type)
	}
	if opts.Area != "" {
		query += " AND m.area = ?"
		args = append(args, opts.Area)
	}

	query += `
		ORDER BY vec_distance_cosine(e.embedding, ?)
		LIMIT ?
	`
	args = append(args, string(embeddingJSON), limit)

	return s.queryMemories(ctx, query, args...)
}

func (s *SQLite) List(ctx context.Context, opts ListOpts) ([]Memory, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, type, area, content, rationale, is_valid, superseded_by, created_at
		FROM memories
		WHERE 1=1
	`
	args := []interface{}{}

	if !opts.IncludeInvalid {
		query += " AND is_valid = TRUE"
	}
	if opts.Type != "" {
		query += " AND type = ?"
		args = append(args, opts.Type)
	}
	if opts.Area != "" {
		query += " AND area = ?"
		args = append(args, opts.Area)
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	return s.queryMemories(ctx, query, args...)
}

func (s *SQLite) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
	query := `UPDATE memories SET is_valid = FALSE`
	args := []interface{}{}

	if supersededBy != nil {
		query += ", superseded_by = ?"
		args = append(args, *supersededBy)
	}

	query += " WHERE id = ?"
	args = append(args, id)

	result, err := s.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("memory with id %d not found", id)
	}

	return nil
}

func (s *SQLite) queryMemories(ctx context.Context, query string, args ...interface{}) ([]Memory, error) {
	rows, err := s.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		var supersededBy sql.NullInt64
		var rationale sql.NullString

		if err := rows.Scan(&m.ID, &m.Type, &m.Area, &m.Content, &rationale, &m.IsValid, &supersededBy, &m.CreatedAt); err != nil {
			return nil, err
		}

		if rationale.Valid {
			m.Rationale = rationale.String
		}
		if supersededBy.Valid {
			m.SupersededBy = &supersededBy.Int64
		}

		memories = append(memories, m)
	}

	return memories, rows.Err()
}
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=1 go test ./internal/storage/... -v`
Expected: PASS

**Step 5: Add remaining tests**

```go
// Append to internal/storage/sqlite_test.go

func TestSQLiteStorage_Search(t *testing.T) {
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	store, err := storage.NewSQLite(f.Name())
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Add a memory first
	mem := storage.Memory{
		Type:    "decision",
		Area:    "auth",
		Content: "Use JWT tokens",
	}
	embedding := make([]float32, 768)
	embedding[0] = 0.5

	_, err = store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Search for it
	results, err := store.Search(ctx, embedding, storage.SearchOpts{Limit: 5})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestSQLiteStorage_List(t *testing.T) {
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	store, err := storage.NewSQLite(f.Name())
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Add memories
	for i := 0; i < 3; i++ {
		mem := storage.Memory{
			Type:    "learning",
			Area:    "api",
			Content: fmt.Sprintf("Learning %d", i),
		}
		embedding := make([]float32, 768)
		_, err = store.Add(ctx, mem, embedding)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	results, err := store.List(ctx, storage.ListOpts{Limit: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestSQLiteStorage_Invalidate(t *testing.T) {
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	store, err := storage.NewSQLite(f.Name())
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	mem := storage.Memory{
		Type:    "decision",
		Area:    "auth",
		Content: "Old decision",
	}
	embedding := make([]float32, 768)

	added, err := store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	err = store.Invalidate(ctx, added.ID, nil)
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	// Should not appear in list (excludes invalid by default)
	results, err := store.List(ctx, storage.ListOpts{Limit: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results after invalidation, got %d", len(results))
	}
}
```

**Step 6: Run all tests**

Run: `CGO_ENABLED=1 go test ./internal/storage/... -v`
Expected: PASS (4 tests)

**Step 7: Commit**

```bash
git add internal/storage/sqlite.go internal/storage/sqlite_test.go
git commit -m "feat(storage): add SQLite implementation of Storage interface"
```

---

## Task 3: Create Embedder Interface

**Files:**
- Create: `internal/embedder/embedder.go`
- Create: `internal/embedder/ollama.go`
- Create: `internal/embedder/ollama_test.go`

**Step 1: Write interface**

```go
// internal/embedder/embedder.go
package embedder

// Embedder generates vector embeddings for text
type Embedder interface {
	// EmbedForStorage creates an embedding optimized for document storage
	EmbedForStorage(text string) ([]float32, error)
	// EmbedForSearch creates an embedding optimized for search queries
	EmbedForSearch(query string) ([]float32, error)
}
```

**Step 2: Move Ollama implementation**

```go
// internal/embedder/ollama.go
package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Ollama implements Embedder using Ollama API
type Ollama struct {
	baseURL string
	model   string
	http    *http.Client
}

type embeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// NewOllama creates a new Ollama embedder
func NewOllama(baseURL, model string) *Ollama {
	return &Ollama{
		baseURL: baseURL,
		model:   model,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (o *Ollama) embed(text string) ([]float32, error) {
	reqBody := embeddingRequest{
		Model:  o.model,
		Prompt: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := o.http.Post(
		fmt.Sprintf("%s/api/embeddings", o.baseURL),
		"application/json",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var embResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return embResp.Embedding, nil
}

func (o *Ollama) EmbedForStorage(text string) ([]float32, error) {
	if o.model == "nomic-embed-text" {
		return o.embed("search_document: " + text)
	}
	return o.embed(text)
}

func (o *Ollama) EmbedForSearch(query string) ([]float32, error) {
	if o.model == "nomic-embed-text" {
		return o.embed("search_query: " + query)
	}
	return o.embed(query)
}
```

**Step 3: Verify it compiles**

Run: `CGO_ENABLED=1 go build ./internal/embedder/...`
Expected: Success

**Step 4: Commit**

```bash
git add internal/embedder/
git commit -m "feat(embedder): add Embedder interface with Ollama implementation"
```

---

## Task 4: Create Service Layer

**Files:**
- Create: `internal/service/service.go`
- Create: `internal/service/service_test.go`

**Step 1: Write failing test with mock**

```go
// internal/service/service_test.go
package service_test

import (
	"context"
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

// mockEmbedder implements embedder.Embedder for testing
type mockEmbedder struct{}

func (m *mockEmbedder) EmbedForStorage(text string) ([]float32, error) {
	return make([]float32, 768), nil
}

func (m *mockEmbedder) EmbedForSearch(query string) ([]float32, error) {
	return make([]float32, 768), nil
}

// mockStorage implements storage.Storage for testing
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

func TestService_Add(t *testing.T) {
	store := &mockStorage{}
	emb := &mockEmbedder{}
	svc := service.New(store, emb)

	ctx := context.Background()
	mem, err := svc.Add(ctx, "decision", "auth", "Use JWT", "Stateless")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if mem.Type != "decision" {
		t.Errorf("expected type 'decision', got %q", mem.Type)
	}
	if mem.Content != "Use JWT" {
		t.Errorf("expected content 'Use JWT', got %q", mem.Content)
	}
}

func TestService_Search(t *testing.T) {
	store := &mockStorage{}
	emb := &mockEmbedder{}
	svc := service.New(store, emb)

	ctx := context.Background()

	// Add a memory first
	_, err := svc.Add(ctx, "decision", "auth", "Use JWT", "")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	results, err := svc.Search(ctx, "jwt tokens", 5, "", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=1 go test ./internal/service/... -v`
Expected: FAIL with "undefined: service.New"

**Step 3: Write Service implementation**

```go
// internal/service/service.go
package service

import (
	"context"
	"fmt"

	"github.com/MereWhiplash/engram-cogitator/internal/embedder"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
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
func (s *Service) Add(ctx context.Context, memType, area, content, rationale string) (*storage.Memory, error) {
	// Build text for embedding
	textToEmbed := fmt.Sprintf("%s: %s", area, content)
	if rationale != "" {
		textToEmbed += " " + rationale
	}

	embedding, err := s.embedder.EmbedForStorage(textToEmbed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	mem := storage.Memory{
		Type:      memType,
		Area:      area,
		Content:   content,
		Rationale: rationale,
	}

	return s.storage.Add(ctx, mem, embedding)
}

// Search finds memories by semantic similarity
func (s *Service) Search(ctx context.Context, query string, limit int, memType, area string) ([]storage.Memory, error) {
	embedding, err := s.embedder.EmbedForSearch(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	opts := storage.SearchOpts{
		Limit: limit,
		Type:  memType,
		Area:  area,
	}

	return s.storage.Search(ctx, embedding, opts)
}

// List returns recent memories
func (s *Service) List(ctx context.Context, limit int, memType, area string, includeInvalid bool) ([]storage.Memory, error) {
	opts := storage.ListOpts{
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
```

**Step 4: Run tests to verify they pass**

Run: `CGO_ENABLED=1 go test ./internal/service/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/service/
git commit -m "feat(service): add Service layer with business logic"
```

---

## Task 5: Update Tool Handlers to Use Service

**Files:**
- Modify: `internal/tools/tools.go`
- Modify: `cmd/server/main.go`

**Step 1: Update tools.go to use Service**

```go
// internal/tools/tools.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Handler holds dependencies for tool handlers
type Handler struct {
	svc *service.Service
}

// AddInput defines the input schema for ec_add
type AddInput struct {
	Type      string `json:"type" jsonschema:"required" jsonschema_description:"Type of memory: decision, learning, or pattern"`
	Area      string `json:"area" jsonschema:"required" jsonschema_description:"Domain area (e.g. auth, permissions, ui, api)"`
	Content   string `json:"content" jsonschema:"required" jsonschema_description:"The actual content to remember"`
	Rationale string `json:"rationale,omitempty" jsonschema_description:"Why this matters or additional context"`
}

// AddOutput defines the output schema for ec_add
type AddOutput struct {
	Memory *storage.Memory `json:"memory"`
}

// SearchInput defines the input schema for ec_search
type SearchInput struct {
	Query string `json:"query" jsonschema:"required" jsonschema_description:"Search query to find relevant memories"`
	Limit int    `json:"limit,omitempty" jsonschema_description:"Maximum number of results (default: 5)"`
	Type  string `json:"type,omitempty" jsonschema_description:"Filter by type (decision, learning, or pattern)"`
	Area  string `json:"area,omitempty" jsonschema_description:"Filter by domain area"`
}

// SearchOutput defines the output schema for ec_search
type SearchOutput struct {
	Memories []storage.Memory `json:"memories"`
}

// InvalidateInput defines the input schema for ec_invalidate
type InvalidateInput struct {
	ID           int64 `json:"id" jsonschema:"required" jsonschema_description:"ID of the memory to invalidate"`
	SupersededBy int64 `json:"superseded_by,omitempty" jsonschema_description:"ID of the memory that supersedes this one"`
}

// InvalidateOutput defines the output schema for ec_invalidate
type InvalidateOutput struct {
	Message string `json:"message"`
}

// ListInput defines the input schema for ec_list
type ListInput struct {
	Limit          int    `json:"limit,omitempty" jsonschema_description:"Maximum number of results (default: 10)"`
	Type           string `json:"type,omitempty" jsonschema_description:"Filter by type (decision, learning, or pattern)"`
	Area           string `json:"area,omitempty" jsonschema_description:"Filter by domain area"`
	IncludeInvalid bool   `json:"include_invalid,omitempty" jsonschema_description:"Include invalidated entries (default: false)"`
}

// ListOutput defines the output schema for ec_list
type ListOutput struct {
	Memories []storage.Memory `json:"memories"`
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}

// Register adds all EC tools to the MCP server
func Register(server *mcp.Server, svc *service.Service) {
	h := &Handler{svc: svc}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ec_add",
		Description: "Add a new memory entry (decision, learning, or pattern)",
	}, h.Add)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ec_search",
		Description: "Search memories by semantic similarity",
	}, h.Search)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ec_invalidate",
		Description: "Invalidate a memory entry (soft delete)",
	}, h.Invalidate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ec_list",
		Description: "List recent memory entries",
	}, h.List)
}

func (h *Handler) Add(ctx context.Context, req *mcp.CallToolRequest, input AddInput) (*mcp.CallToolResult, AddOutput, error) {
	if input.Type == "" || input.Area == "" || input.Content == "" {
		return errorResult("type, area, and content are required"), AddOutput{}, nil
	}

	memory, err := h.svc.Add(ctx, input.Type, input.Area, input.Content, input.Rationale)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to store memory: %v", err)), AddOutput{}, nil
	}

	result, _ := json.MarshalIndent(memory, "", "  ")
	return textResult(fmt.Sprintf("Memory added successfully:\n%s", string(result))), AddOutput{Memory: memory}, nil
}

func (h *Handler) Search(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
	if input.Query == "" {
		return errorResult("query is required"), SearchOutput{}, nil
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 5
	}

	memories, err := h.svc.Search(ctx, input.Query, limit, input.Type, input.Area)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to search: %v", err)), SearchOutput{}, nil
	}

	if len(memories) == 0 {
		return textResult("No matching memories found."), SearchOutput{Memories: []storage.Memory{}}, nil
	}

	result, _ := json.MarshalIndent(memories, "", "  ")
	return textResult(string(result)), SearchOutput{Memories: memories}, nil
}

func (h *Handler) Invalidate(ctx context.Context, req *mcp.CallToolRequest, input InvalidateInput) (*mcp.CallToolResult, InvalidateOutput, error) {
	if input.ID == 0 {
		return errorResult("id is required"), InvalidateOutput{}, nil
	}

	var supersededBy *int64
	if input.SupersededBy > 0 {
		supersededBy = &input.SupersededBy
	}

	if err := h.svc.Invalidate(ctx, input.ID, supersededBy); err != nil {
		return errorResult(fmt.Sprintf("failed to invalidate: %v", err)), InvalidateOutput{}, nil
	}

	msg := fmt.Sprintf("Memory %d has been invalidated.", input.ID)
	if supersededBy != nil {
		msg += fmt.Sprintf(" Superseded by memory %d.", *supersededBy)
	}

	return textResult(msg), InvalidateOutput{Message: msg}, nil
}

func (h *Handler) List(ctx context.Context, req *mcp.CallToolRequest, input ListInput) (*mcp.CallToolResult, ListOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	memories, err := h.svc.List(ctx, limit, input.Type, input.Area, input.IncludeInvalid)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to list: %v", err)), ListOutput{}, nil
	}

	if len(memories) == 0 {
		return textResult("No memories found."), ListOutput{Memories: []storage.Memory{}}, nil
	}

	result, _ := json.MarshalIndent(memories, "", "  ")
	return textResult(string(result)), ListOutput{Memories: memories}, nil
}
```

**Step 2: Update main.go to use new packages**

```go
// cmd/server/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MereWhiplash/engram-cogitator/internal/embedder"
	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
	"github.com/MereWhiplash/engram-cogitator/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	dbPath := flag.String("db-path", "/data/memory.db", "Path to SQLite database")
	ollamaURL := flag.String("ollama-url", "http://ollama:11434", "Ollama API URL")
	embeddingModel := flag.String("embedding-model", "nomic-embed-text", "Ollama embedding model")

	// CLI mode flags
	listFlag := flag.Bool("list", false, "List recent memories (CLI mode)")
	limitFlag := flag.Int("limit", 5, "Limit for list operation")

	flag.Parse()

	// CLI mode - list memories
	if *listFlag {
		if err := runList(*dbPath, *limitFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Initialize storage
	store, err := storage.NewSQLite(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize embedder
	emb := embedder.NewOllama(*ollamaURL, *embeddingModel)

	// Create service
	svc := service.New(store, emb)

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "engram-cogitator",
		Version: "1.0.0",
	}, nil)

	// Register tools
	tools.Register(server, svc)

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Start server with stdio transport
	log.Println("Starting Engram Cogitator MCP server...")
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func runList(dbPath string, limit int) error {
	store, err := storage.NewSQLite(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer store.Close()

	ctx := context.Background()
	memories, err := store.List(ctx, storage.ListOpts{Limit: limit})
	if err != nil {
		return fmt.Errorf("failed to list memories: %w", err)
	}

	if len(memories) == 0 {
		return nil
	}

	for _, m := range memories {
		fmt.Printf("[%s/%s] %s\n", m.Type, m.Area, m.Content)
	}
	return nil
}
```

**Step 3: Remove old db and embed packages**

Run: `rm -rf internal/db internal/embed`

**Step 4: Run all tests**

Run: `CGO_ENABLED=1 go test ./... -v`
Expected: PASS

**Step 5: Verify build**

Run: `CGO_ENABLED=1 go build ./cmd/server`
Expected: Success

**Step 6: Commit**

```bash
git add -A
git commit -m "refactor: use Service layer with Storage and Embedder interfaces

- tools.Handler now depends on service.Service
- main.go wires up storage.SQLite and embedder.Ollama
- Removed old internal/db and internal/embed packages
- All existing behavior preserved"
```

---

## Task 6: Clean Up - Remove Old Packages

**Files:**
- Delete: `internal/db/` (replaced by `internal/storage/`)
- Delete: `internal/embed/` (replaced by `internal/embedder/`)

**Step 1: Verify old packages are gone**

Run: `ls internal/`
Expected: `embedder  service  storage  tools`

**Step 2: Run full test suite**

Run: `CGO_ENABLED=1 go test ./... -v`
Expected: All tests pass

**Step 3: Build final binary**

Run: `CGO_ENABLED=1 go build ./cmd/server`
Expected: Success

**Step 4: Final commit (if any cleanup needed)**

```bash
git status
# If clean, done. If not:
git add -A
git commit -m "chore: cleanup after refactor"
```

---

## Summary

After Phase 1, you'll have:

```
internal/
  storage/
    storage.go      # Storage interface + types
    sqlite.go       # SQLite implementation
    sqlite_test.go  # Tests
  embedder/
    embedder.go     # Embedder interface
    ollama.go       # Ollama implementation
  service/
    service.go      # Business logic layer
    service_test.go # Tests with mocks
  tools/
    tools.go        # MCP handlers (uses Service)
```

**Next phase:** Add Postgres and MongoDB implementations of Storage interface.
