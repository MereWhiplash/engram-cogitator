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
	if err := mem.Type.Validate(); err != nil {
		return nil, err
	}

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
		var memType string
		var supersededBy sql.NullInt64
		var rationale sql.NullString

		if err := rows.Scan(&m.ID, &memType, &m.Area, &m.Content, &rationale, &m.IsValid, &supersededBy, &m.CreatedAt); err != nil {
			return nil, err
		}

		m.Type = MemoryType(memType)
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
