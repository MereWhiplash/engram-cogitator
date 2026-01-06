# Phase 2: Add Storage Backends - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add Postgres and MongoDB implementations of the Storage interface, with full test coverage.

**Architecture:** Storage interface from Phase 1 gets two new implementations. Config flag selects which backend. Postgres uses pgvector for similarity search, MongoDB uses Atlas Vector Search.

**Tech Stack:** Go 1.23, pgx (Postgres driver), mongo-go-driver, golang-migrate

**Prerequisites:** Phase 1 complete (Storage interface exists)

---

## Task 1: Add Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add Postgres and MongoDB drivers**

Run:
```bash
go get github.com/jackc/pgx/v5
go get github.com/pgvector/pgvector-go
go get go.mongodb.org/mongo-driver/mongo
go get github.com/golang-migrate/migrate/v4
```

**Step 2: Verify go.mod updated**

Run: `cat go.mod`
Expected: New dependencies listed

**Step 3: Tidy**

Run: `go mod tidy`

**Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add postgres and mongodb driver dependencies"
```

---

## Task 2: Create Postgres Storage Implementation

**Files:**
- Create: `internal/storage/postgres.go`
- Create: `internal/storage/postgres_test.go`

**Step 1: Write failing test**

```go
// internal/storage/postgres_test.go
package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

func TestPostgresStorage_Add(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set, skipping Postgres tests")
	}

	ctx := context.Background()
	store, err := storage.NewPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	mem := storage.Memory{
		Type:        "decision",
		Area:        "auth",
		Content:     "Use JWT tokens",
		Rationale:   "Stateless auth",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Repo:        "testorg/testrepo",
	}
	embedding := make([]float32, 768)
	embedding[0] = 0.5

	result, err := store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if result.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if result.AuthorEmail != "test@example.com" {
		t.Errorf("expected author email 'test@example.com', got %q", result.AuthorEmail)
	}
}

func TestPostgresStorage_Search(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set, skipping Postgres tests")
	}

	ctx := context.Background()
	store, err := storage.NewPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	// Add a memory
	mem := storage.Memory{
		Type:        "decision",
		Area:        "auth",
		Content:     "Use JWT tokens",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Repo:        "testorg/testrepo",
	}
	embedding := make([]float32, 768)
	embedding[0] = 0.5

	_, err = store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Search - should find it
	results, err := store.Search(ctx, embedding, storage.SearchOpts{Limit: 5})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) < 1 {
		t.Error("expected at least 1 result")
	}
}

func TestPostgresStorage_SearchByRepo(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set, skipping Postgres tests")
	}

	ctx := context.Background()
	store, err := storage.NewPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	embedding := make([]float32, 768)

	// Add memories to different repos
	mem1 := storage.Memory{
		Type:        "decision",
		Area:        "auth",
		Content:     "Repo A decision",
		AuthorName:  "User A",
		AuthorEmail: "a@example.com",
		Repo:        "org/repo-a",
	}
	mem2 := storage.Memory{
		Type:        "decision",
		Area:        "auth",
		Content:     "Repo B decision",
		AuthorName:  "User B",
		AuthorEmail: "b@example.com",
		Repo:        "org/repo-b",
	}

	_, _ = store.Add(ctx, mem1, embedding)
	_, _ = store.Add(ctx, mem2, embedding)

	// Search scoped to repo-a
	results, err := store.Search(ctx, embedding, storage.SearchOpts{
		Limit: 10,
		Repo:  "org/repo-a",
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	for _, r := range results {
		if r.Repo != "org/repo-a" {
			t.Errorf("expected repo 'org/repo-a', got %q", r.Repo)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=1 go test ./internal/storage/... -v -run Postgres`
Expected: FAIL with "undefined: storage.NewPostgres"

**Step 3: Write Postgres implementation**

```go
// internal/storage/postgres.go
package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// Postgres implements Storage using PostgreSQL with pgvector
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres creates a new Postgres storage
func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	p := &Postgres{pool: pool}
	if err := p.initSchema(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return p, nil
}

func (p *Postgres) initSchema(ctx context.Context) error {
	schema := `
		CREATE EXTENSION IF NOT EXISTS vector;

		CREATE TABLE IF NOT EXISTS memories (
			id SERIAL PRIMARY KEY,
			type TEXT NOT NULL CHECK(type IN ('decision', 'learning', 'pattern')),
			area TEXT NOT NULL,
			content TEXT NOT NULL,
			rationale TEXT,
			is_valid BOOLEAN NOT NULL DEFAULT TRUE,
			superseded_by INTEGER REFERENCES memories(id),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			author_name TEXT NOT NULL DEFAULT '',
			author_email TEXT NOT NULL DEFAULT '',
			repo TEXT NOT NULL DEFAULT ''
		);

		CREATE TABLE IF NOT EXISTS memory_embeddings (
			memory_id INTEGER PRIMARY KEY REFERENCES memories(id) ON DELETE CASCADE,
			embedding vector(768)
		);

		CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
		CREATE INDEX IF NOT EXISTS idx_memories_area ON memories(area);
		CREATE INDEX IF NOT EXISTS idx_memories_is_valid ON memories(is_valid);
		CREATE INDEX IF NOT EXISTS idx_memories_repo ON memories(repo);
		CREATE INDEX IF NOT EXISTS idx_memories_author ON memories(author_email);
		CREATE INDEX IF NOT EXISTS idx_memory_embeddings_vector ON memory_embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
	`
	_, err := p.pool.Exec(ctx, schema)
	return err
}

func (p *Postgres) Close() error {
	p.pool.Close()
	return nil
}

func (p *Postgres) Add(ctx context.Context, mem Memory, embedding []float32) (*Memory, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var id int64
	var createdAt time.Time
	err = tx.QueryRow(ctx,
		`INSERT INTO memories (type, area, content, rationale, author_name, author_email, repo)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at`,
		mem.Type, mem.Area, mem.Content, mem.Rationale,
		mem.AuthorName, mem.AuthorEmail, mem.Repo,
	).Scan(&id, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert memory: %w", err)
	}

	vec := pgvector.NewVector(embedding)
	_, err = tx.Exec(ctx,
		`INSERT INTO memory_embeddings (memory_id, embedding) VALUES ($1, $2)`,
		id, vec,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert embedding: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &Memory{
		ID:          id,
		Type:        mem.Type,
		Area:        mem.Area,
		Content:     mem.Content,
		Rationale:   mem.Rationale,
		IsValid:     true,
		CreatedAt:   createdAt,
		AuthorName:  mem.AuthorName,
		AuthorEmail: mem.AuthorEmail,
		Repo:        mem.Repo,
	}, nil
}

func (p *Postgres) Search(ctx context.Context, embedding []float32, opts SearchOpts) ([]Memory, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 5
	}

	vec := pgvector.NewVector(embedding)

	query := `
		SELECT m.id, m.type, m.area, m.content, m.rationale, m.is_valid,
		       m.superseded_by, m.created_at, m.author_name, m.author_email, m.repo
		FROM memories m
		JOIN memory_embeddings e ON m.id = e.memory_id
		WHERE m.is_valid = TRUE
	`
	args := []interface{}{vec}
	argNum := 2

	if opts.Type != "" {
		query += fmt.Sprintf(" AND m.type = $%d", argNum)
		args = append(args, opts.Type)
		argNum++
	}
	if opts.Area != "" {
		query += fmt.Sprintf(" AND m.area = $%d", argNum)
		args = append(args, opts.Area)
		argNum++
	}
	if opts.Repo != "" {
		query += fmt.Sprintf(" AND m.repo = $%d", argNum)
		args = append(args, opts.Repo)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY e.embedding <=> $1 LIMIT $%d", argNum)
	args = append(args, limit)

	return p.queryMemories(ctx, query, args...)
}

func (p *Postgres) List(ctx context.Context, opts ListOpts) ([]Memory, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, type, area, content, rationale, is_valid,
		       superseded_by, created_at, author_name, author_email, repo
		FROM memories
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if !opts.IncludeInvalid {
		query += " AND is_valid = TRUE"
	}
	if opts.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argNum)
		args = append(args, opts.Type)
		argNum++
	}
	if opts.Area != "" {
		query += fmt.Sprintf(" AND area = $%d", argNum)
		args = append(args, opts.Area)
		argNum++
	}
	if opts.Repo != "" {
		query += fmt.Sprintf(" AND repo = $%d", argNum)
		args = append(args, opts.Repo)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argNum)
	args = append(args, limit)

	return p.queryMemories(ctx, query, args...)
}

func (p *Postgres) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
	var result pgx.Rows
	var err error

	if supersededBy != nil {
		result, err = p.pool.Query(ctx,
			`UPDATE memories SET is_valid = FALSE, superseded_by = $1 WHERE id = $2`,
			*supersededBy, id,
		)
	} else {
		result, err = p.pool.Query(ctx,
			`UPDATE memories SET is_valid = FALSE WHERE id = $1`,
			id,
		)
	}
	if err != nil {
		return err
	}
	result.Close()

	return nil
}

func (p *Postgres) queryMemories(ctx context.Context, query string, args ...interface{}) ([]Memory, error) {
	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		var supersededBy *int64
		var rationale *string

		err := rows.Scan(
			&m.ID, &m.Type, &m.Area, &m.Content, &rationale, &m.IsValid,
			&supersededBy, &m.CreatedAt, &m.AuthorName, &m.AuthorEmail, &m.Repo,
		)
		if err != nil {
			return nil, err
		}

		if rationale != nil {
			m.Rationale = *rationale
		}
		m.SupersededBy = supersededBy

		memories = append(memories, m)
	}

	return memories, rows.Err()
}
```

**Step 4: Run tests (will skip without TEST_POSTGRES_DSN)**

Run: `CGO_ENABLED=1 go test ./internal/storage/... -v`
Expected: Postgres tests SKIP, SQLite tests PASS

**Step 5: Commit**

```bash
git add internal/storage/postgres.go internal/storage/postgres_test.go
git commit -m "feat(storage): add Postgres implementation with pgvector"
```

---

## Task 3: Create MongoDB Storage Implementation

**Files:**
- Create: `internal/storage/mongodb.go`
- Create: `internal/storage/mongodb_test.go`

**Step 1: Write failing test**

```go
// internal/storage/mongodb_test.go
package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

func TestMongoDBStorage_Add(t *testing.T) {
	uri := os.Getenv("TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("TEST_MONGODB_URI not set, skipping MongoDB tests")
	}

	ctx := context.Background()
	store, err := storage.NewMongoDB(ctx, uri, "engram_test")
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	mem := storage.Memory{
		Type:        "decision",
		Area:        "auth",
		Content:     "Use JWT tokens",
		Rationale:   "Stateless auth",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Repo:        "testorg/testrepo",
	}
	embedding := make([]float32, 768)
	embedding[0] = 0.5

	result, err := store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if result.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if result.AuthorEmail != "test@example.com" {
		t.Errorf("expected author email 'test@example.com', got %q", result.AuthorEmail)
	}
}

func TestMongoDBStorage_Search(t *testing.T) {
	uri := os.Getenv("TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("TEST_MONGODB_URI not set, skipping MongoDB tests")
	}

	ctx := context.Background()
	store, err := storage.NewMongoDB(ctx, uri, "engram_test")
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	mem := storage.Memory{
		Type:        "learning",
		Area:        "api",
		Content:     "Rate limiting patterns",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Repo:        "testorg/testrepo",
	}
	embedding := make([]float32, 768)
	embedding[0] = 0.8

	_, err = store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	results, err := store.Search(ctx, embedding, storage.SearchOpts{Limit: 5})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) < 1 {
		t.Error("expected at least 1 result")
	}
}

func TestMongoDBStorage_List(t *testing.T) {
	uri := os.Getenv("TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("TEST_MONGODB_URI not set, skipping MongoDB tests")
	}

	ctx := context.Background()
	store, err := storage.NewMongoDB(ctx, uri, "engram_test")
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	mem := storage.Memory{
		Type:        "pattern",
		Area:        "testing",
		Content:     "Always use table-driven tests",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Repo:        "testorg/testrepo",
	}
	embedding := make([]float32, 768)

	_, err = store.Add(ctx, mem, embedding)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	results, err := store.List(ctx, storage.ListOpts{Limit: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(results) < 1 {
		t.Error("expected at least 1 result")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=1 go test ./internal/storage/... -v -run MongoDB`
Expected: FAIL with "undefined: storage.NewMongoDB"

**Step 3: Write MongoDB implementation**

```go
// internal/storage/mongodb.go
package storage

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB implements Storage using MongoDB with Atlas Vector Search
type MongoDB struct {
	client     *mongo.Client
	db         *mongo.Database
	memories   *mongo.Collection
	idCounter  int64
}

// memoryDoc is the MongoDB document structure
type memoryDoc struct {
	ID           int64     `bson:"_id"`
	Type         string    `bson:"type"`
	Area         string    `bson:"area"`
	Content      string    `bson:"content"`
	Rationale    string    `bson:"rationale,omitempty"`
	IsValid      bool      `bson:"is_valid"`
	SupersededBy *int64    `bson:"superseded_by,omitempty"`
	CreatedAt    time.Time `bson:"created_at"`
	Author       struct {
		Name  string `bson:"name"`
		Email string `bson:"email"`
	} `bson:"author"`
	Repo      string    `bson:"repo"`
	Embedding []float32 `bson:"embedding"`
}

// NewMongoDB creates a new MongoDB storage
func NewMongoDB(ctx context.Context, uri, database string) (*MongoDB, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	db := client.Database(database)
	memories := db.Collection("memories")

	m := &MongoDB{
		client:   client,
		db:       db,
		memories: memories,
	}

	if err := m.initIndexes(ctx); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	// Initialize ID counter from max existing ID
	if err := m.initIDCounter(ctx); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to init id counter: %w", err)
	}

	return m, nil
}

func (m *MongoDB) initIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "type", Value: 1}}},
		{Keys: bson.D{{Key: "area", Value: 1}}},
		{Keys: bson.D{{Key: "is_valid", Value: 1}}},
		{Keys: bson.D{{Key: "repo", Value: 1}}},
		{Keys: bson.D{{Key: "author.email", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	}

	_, err := m.memories.Indexes().CreateMany(ctx, indexes)
	return err
}

func (m *MongoDB) initIDCounter(ctx context.Context) error {
	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	var doc memoryDoc
	err := m.memories.FindOne(ctx, bson.D{}, opts).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		m.idCounter = 0
		return nil
	}
	if err != nil {
		return err
	}
	m.idCounter = doc.ID
	return nil
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

func (m *MongoDB) Add(ctx context.Context, mem Memory, embedding []float32) (*Memory, error) {
	id := atomic.AddInt64(&m.idCounter, 1)
	now := time.Now()

	doc := memoryDoc{
		ID:        id,
		Type:      mem.Type,
		Area:      mem.Area,
		Content:   mem.Content,
		Rationale: mem.Rationale,
		IsValid:   true,
		CreatedAt: now,
		Repo:      mem.Repo,
		Embedding: embedding,
	}
	doc.Author.Name = mem.AuthorName
	doc.Author.Email = mem.AuthorEmail

	_, err := m.memories.InsertOne(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to insert memory: %w", err)
	}

	return &Memory{
		ID:          id,
		Type:        mem.Type,
		Area:        mem.Area,
		Content:     mem.Content,
		Rationale:   mem.Rationale,
		IsValid:     true,
		CreatedAt:   now,
		AuthorName:  mem.AuthorName,
		AuthorEmail: mem.AuthorEmail,
		Repo:        mem.Repo,
	}, nil
}

func (m *MongoDB) Search(ctx context.Context, embedding []float32, opts SearchOpts) ([]Memory, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 5
	}

	// Build filter
	filter := bson.D{{Key: "is_valid", Value: true}}
	if opts.Type != "" {
		filter = append(filter, bson.E{Key: "type", Value: opts.Type})
	}
	if opts.Area != "" {
		filter = append(filter, bson.E{Key: "area", Value: opts.Area})
	}
	if opts.Repo != "" {
		filter = append(filter, bson.E{Key: "repo", Value: opts.Repo})
	}

	// Atlas Vector Search pipeline
	// Note: This requires an Atlas Vector Search index named "embedding_index"
	// For non-Atlas deployments, falls back to regular query (no vector search)
	pipeline := mongo.Pipeline{
		{{Key: "$vectorSearch", Value: bson.D{
			{Key: "index", Value: "embedding_index"},
			{Key: "path", Value: "embedding"},
			{Key: "queryVector", Value: embedding},
			{Key: "numCandidates", Value: limit * 10},
			{Key: "limit", Value: limit},
			{Key: "filter", Value: filter},
		}}},
	}

	cursor, err := m.memories.Aggregate(ctx, pipeline)
	if err != nil {
		// Fallback to regular query if vector search not available
		return m.listFallback(ctx, opts)
	}
	defer cursor.Close(ctx)

	return m.cursorToMemories(ctx, cursor)
}

func (m *MongoDB) listFallback(ctx context.Context, opts SearchOpts) ([]Memory, error) {
	listOpts := ListOpts{
		Limit: opts.Limit,
		Type:  opts.Type,
		Area:  opts.Area,
		Repo:  opts.Repo,
	}
	return m.List(ctx, listOpts)
}

func (m *MongoDB) List(ctx context.Context, opts ListOpts) ([]Memory, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}

	filter := bson.D{}
	if !opts.IncludeInvalid {
		filter = append(filter, bson.E{Key: "is_valid", Value: true})
	}
	if opts.Type != "" {
		filter = append(filter, bson.E{Key: "type", Value: opts.Type})
	}
	if opts.Area != "" {
		filter = append(filter, bson.E{Key: "area", Value: opts.Area})
	}
	if opts.Repo != "" {
		filter = append(filter, bson.E{Key: "repo", Value: opts.Repo})
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := m.memories.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	return m.cursorToMemories(ctx, cursor)
}

func (m *MongoDB) Invalidate(ctx context.Context, id int64, supersededBy *int64) error {
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "is_valid", Value: false}}}}
	if supersededBy != nil {
		update = bson.D{{Key: "$set", Value: bson.D{
			{Key: "is_valid", Value: false},
			{Key: "superseded_by", Value: *supersededBy},
		}}}
	}

	result, err := m.memories.UpdateOne(ctx, bson.D{{Key: "_id", Value: id}}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("memory with id %d not found", id)
	}

	return nil
}

func (m *MongoDB) cursorToMemories(ctx context.Context, cursor *mongo.Cursor) ([]Memory, error) {
	var memories []Memory
	for cursor.Next(ctx) {
		var doc memoryDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		memories = append(memories, Memory{
			ID:           doc.ID,
			Type:         doc.Type,
			Area:         doc.Area,
			Content:      doc.Content,
			Rationale:    doc.Rationale,
			IsValid:      doc.IsValid,
			SupersededBy: doc.SupersededBy,
			CreatedAt:    doc.CreatedAt,
			AuthorName:   doc.Author.Name,
			AuthorEmail:  doc.Author.Email,
			Repo:         doc.Repo,
		})
	}

	return memories, cursor.Err()
}
```

**Step 4: Run tests**

Run: `CGO_ENABLED=1 go test ./internal/storage/... -v`
Expected: MongoDB tests SKIP, others PASS

**Step 5: Commit**

```bash
git add internal/storage/mongodb.go internal/storage/mongodb_test.go
git commit -m "feat(storage): add MongoDB implementation with Atlas Vector Search"
```

---

## Task 4: Add Storage Factory Function

**Files:**
- Create: `internal/storage/factory.go`
- Create: `internal/storage/factory_test.go`

**Step 1: Write factory function**

```go
// internal/storage/factory.go
package storage

import (
	"context"
	"fmt"
)

// Config holds storage configuration
type Config struct {
	Driver   string // "sqlite", "postgres", "mongodb"

	// SQLite
	SQLitePath string

	// Postgres
	PostgresDSN string

	// MongoDB
	MongoDBURI      string
	MongoDBDatabase string
}

// New creates a Storage implementation based on config
func New(ctx context.Context, cfg Config) (Storage, error) {
	switch cfg.Driver {
	case "sqlite":
		if cfg.SQLitePath == "" {
			return nil, fmt.Errorf("sqlite path is required")
		}
		return NewSQLite(cfg.SQLitePath)

	case "postgres":
		if cfg.PostgresDSN == "" {
			return nil, fmt.Errorf("postgres DSN is required")
		}
		return NewPostgres(ctx, cfg.PostgresDSN)

	case "mongodb":
		if cfg.MongoDBURI == "" {
			return nil, fmt.Errorf("mongodb URI is required")
		}
		if cfg.MongoDBDatabase == "" {
			cfg.MongoDBDatabase = "engram"
		}
		return NewMongoDB(ctx, cfg.MongoDBURI, cfg.MongoDBDatabase)

	default:
		return nil, fmt.Errorf("unknown storage driver: %s", cfg.Driver)
	}
}
```

**Step 2: Write test**

```go
// internal/storage/factory_test.go
package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

func TestNew_SQLite(t *testing.T) {
	f, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	ctx := context.Background()
	store, err := storage.New(ctx, storage.Config{
		Driver:     "sqlite",
		SQLitePath: f.Name(),
	})
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	// Verify it works
	mem := storage.Memory{
		Type:    "decision",
		Area:    "test",
		Content: "Test content",
	}
	_, err = store.Add(ctx, mem, make([]float32, 768))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
}

func TestNew_UnknownDriver(t *testing.T) {
	ctx := context.Background()
	_, err := storage.New(ctx, storage.Config{
		Driver: "unknown",
	})
	if err == nil {
		t.Error("expected error for unknown driver")
	}
}
```

**Step 3: Run tests**

Run: `CGO_ENABLED=1 go test ./internal/storage/... -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/storage/factory.go internal/storage/factory_test.go
git commit -m "feat(storage): add factory function for storage creation"
```

---

## Task 5: Update main.go to Use Storage Config

**Files:**
- Modify: `cmd/server/main.go`

**Step 1: Update main.go with storage driver flag**

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
	// Storage flags
	storageDriver := flag.String("storage-driver", "sqlite", "Storage driver: sqlite, postgres, mongodb")
	dbPath := flag.String("db-path", "/data/memory.db", "Path to SQLite database (sqlite driver)")
	postgresDSN := flag.String("postgres-dsn", "", "PostgreSQL connection string (postgres driver)")
	mongoURI := flag.String("mongodb-uri", "", "MongoDB connection URI (mongodb driver)")
	mongoDatabase := flag.String("mongodb-database", "engram", "MongoDB database name (mongodb driver)")

	// Embedder flags
	ollamaURL := flag.String("ollama-url", "http://ollama:11434", "Ollama API URL")
	embeddingModel := flag.String("embedding-model", "nomic-embed-text", "Ollama embedding model")

	// CLI mode flags
	listFlag := flag.Bool("list", false, "List recent memories (CLI mode)")
	limitFlag := flag.Int("limit", 5, "Limit for list operation")

	flag.Parse()

	ctx := context.Background()

	// Build storage config
	cfg := storage.Config{
		Driver:          *storageDriver,
		SQLitePath:      *dbPath,
		PostgresDSN:     *postgresDSN,
		MongoDBURI:      *mongoURI,
		MongoDBDatabase: *mongoDatabase,
	}

	// CLI mode - list memories
	if *listFlag {
		if err := runList(ctx, cfg, *limitFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Initialize storage
	store, err := storage.New(ctx, cfg)
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
	ctx, cancel := context.WithCancel(ctx)
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

func runList(ctx context.Context, cfg storage.Config, limit int) error {
	store, err := storage.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer store.Close()

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

**Step 2: Verify build**

Run: `CGO_ENABLED=1 go build ./cmd/server`
Expected: Success

**Step 3: Verify all tests pass**

Run: `CGO_ENABLED=1 go test ./... -v`
Expected: PASS

**Step 4: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat: add storage driver selection via CLI flags

Supports --storage-driver sqlite|postgres|mongodb with
appropriate connection flags for each backend."
```

---

## Summary

After Phase 2, you'll have:

```
internal/storage/
  storage.go       # Storage interface
  sqlite.go        # SQLite implementation (solo mode)
  postgres.go      # Postgres + pgvector implementation
  mongodb.go       # MongoDB + Atlas Vector Search implementation
  factory.go       # Factory function to create storage by config
  *_test.go        # Tests for each
```

**CLI usage:**
```bash
# SQLite (default, solo mode)
./server --storage-driver sqlite --db-path ./memory.db

# Postgres (team mode)
./server --storage-driver postgres --postgres-dsn "postgres://user:pass@host/db"

# MongoDB (team mode)
./server --storage-driver mongodb --mongodb-uri "mongodb+srv://..." --mongodb-database engram
```

**Next phase:** Build Central API (HTTP server).
