package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
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
}

// DB wraps the SQLite database with vector search capabilities
type DB struct {
	conn *sql.DB
}

// New creates a new database connection and initializes the schema
func New(path string) (*DB, error) {
	sqlite_vec.Auto()

	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.conn.Close()
}

// initSchema creates the necessary tables if they don't exist
func (d *DB) initSchema() error {
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

	_, err := d.conn.Exec(schema)
	return err
}

// Add inserts a new memory entry with its embedding
func (d *DB) Add(memType, area, content, rationale string, embedding []float32) (*Memory, error) {
	tx, err := d.conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert memory
	result, err := tx.Exec(
		`INSERT INTO memories (type, area, content, rationale) VALUES (?, ?, ?, ?)`,
		memType, area, content, rationale,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert memory: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Insert embedding
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding: %w", err)
	}

	_, err = tx.Exec(
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
		Type:      memType,
		Area:      area,
		Content:   content,
		Rationale: rationale,
		IsValid:   true,
		CreatedAt: time.Now(),
	}, nil
}

// Search finds memories by semantic similarity
func (d *DB) Search(embedding []float32, limit int, memType, area string) ([]Memory, error) {
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

	if memType != "" {
		query += " AND m.type = ?"
		args = append(args, memType)
	}
	if area != "" {
		query += " AND m.area = ?"
		args = append(args, area)
	}

	query += `
		ORDER BY vec_distance_cosine(e.embedding, ?)
		LIMIT ?
	`
	args = append(args, string(embeddingJSON), limit)

	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
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

// List returns recent memories with optional filtering
func (d *DB) List(limit int, memType, area string, includeInvalid bool) ([]Memory, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, type, area, content, rationale, is_valid, superseded_by, created_at
		FROM memories
		WHERE 1=1
	`
	args := []interface{}{}

	if !includeInvalid {
		query += " AND is_valid = TRUE"
	}
	if memType != "" {
		query += " AND type = ?"
		args = append(args, memType)
	}
	if area != "" {
		query += " AND area = ?"
		args = append(args, area)
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := d.conn.Query(query, args...)
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

// Invalidate marks a memory as invalid
func (d *DB) Invalidate(id int64, supersededBy *int64) error {
	query := `UPDATE memories SET is_valid = FALSE`
	args := []interface{}{}

	if supersededBy != nil {
		query += ", superseded_by = ?"
		args = append(args, *supersededBy)
	}

	query += " WHERE id = ?"
	args = append(args, id)

	result, err := d.conn.Exec(query, args...)
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
