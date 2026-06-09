package main

import (
	"fmt"

	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

// buildStorageConfig assembles storage.Config from flags and validates it.
// Unlike the old main(), SQLite is permitted (single shared local api process).
func buildStorageConfig(driver, dbPath, postgresDSN, mongoURI, mongoDatabase string) (storage.Config, error) {
	cfg := storage.Config{
		Driver:          driver,
		SQLitePath:      dbPath,
		PostgresDSN:     postgresDSN,
		MongoDBURI:      mongoURI,
		MongoDBDatabase: mongoDatabase,
	}
	switch driver {
	case "sqlite":
		if dbPath == "" {
			return storage.Config{}, fmt.Errorf("sqlite driver requires --db-path")
		}
	case "postgres":
		if postgresDSN == "" {
			return storage.Config{}, fmt.Errorf("postgres driver requires --postgres-dsn")
		}
	case "mongodb":
		if mongoURI == "" {
			return storage.Config{}, fmt.Errorf("mongodb driver requires --mongodb-uri")
		}
	default:
		return storage.Config{}, fmt.Errorf("unknown storage driver: %s", driver)
	}
	return cfg, nil
}
