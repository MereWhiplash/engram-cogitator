package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"github.com/MereWhiplash/engram-cogitator/internal/types"
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

		CREATE INDEX IF NOT EXISTS idx_embeddings_vector
		ON memory_embeddings USING hnsw (embedding vector_cosine_ops);
	`
	_, err := p.pool.Exec(ctx, schema)
	return err
}

func (p *Postgres) Close() error {
	p.pool.Close()
	return nil
}

func (p *Postgres) Add(ctx context.Context, mem types.Memory, embedding []float32) (*types.Memory, error) {
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

	return &types.Memory{
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

func (p *Postgres) Search(ctx context.Context, embedding []float32, opts types.SearchOpts) ([]types.Memory, error) {
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

func (p *Postgres) List(ctx context.Context, opts types.ListOpts) ([]types.Memory, error) {
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
	var result pgconn.CommandTag
	var err error

	if supersededBy != nil {
		result, err = p.pool.Exec(ctx,
			`UPDATE memories SET is_valid = FALSE, superseded_by = $1 WHERE id = $2`,
			*supersededBy, id,
		)
	} else {
		result, err = p.pool.Exec(ctx,
			`UPDATE memories SET is_valid = FALSE WHERE id = $1`,
			id,
		)
	}
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("memory with id %d not found", id)
	}

	return nil
}

func (p *Postgres) queryMemories(ctx context.Context, query string, args ...interface{}) ([]types.Memory, error) {
	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []types.Memory
	for rows.Next() {
		var m types.Memory
		var memType string
		var supersededBy *int64
		var rationale *string

		err := rows.Scan(
			&m.ID, &memType, &m.Area, &m.Content, &rationale, &m.IsValid,
			&supersededBy, &m.CreatedAt, &m.AuthorName, &m.AuthorEmail, &m.Repo,
		)
		if err != nil {
			return nil, err
		}

		m.Type = types.MemoryType(memType)
		if rationale != nil {
			m.Rationale = *rationale
		}
		m.SupersededBy = supersededBy

		memories = append(memories, m)
	}

	return memories, rows.Err()
}
