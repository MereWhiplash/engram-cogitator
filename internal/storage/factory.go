package storage

import (
	"context"
	"fmt"
)

// Config holds storage configuration
type Config struct {
	Driver string // "sqlite", "postgres", "mongodb"

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
