package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MereWhiplash/engram-cogitator/internal/embedder"
	"github.com/MereWhiplash/engram-cogitator/internal/gitinfo"
	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
	"github.com/MereWhiplash/engram-cogitator/internal/tools"
	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

// version is set by goreleaser via ldflags
var version = "dev"

// defaultDBPath returns the default database path (~/.engram/memory.db)
func defaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".engram/memory.db" // fallback to local
	}
	return filepath.Join(home, ".engram", "memory.db")
}

// expandPath expands ~ to home directory in paths
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

func main() {
	// Storage flags
	storageDriver := flag.String("storage-driver", "sqlite", "Storage driver: sqlite, postgres, mongodb")
	dbPath := flag.String("db-path", defaultDBPath(), "Path to SQLite database (sqlite driver)")
	postgresDSN := flag.String("postgres-dsn", "", "PostgreSQL connection string (postgres driver)")
	mongoURI := flag.String("mongodb-uri", "", "MongoDB connection URI (mongodb driver)")
	mongoDatabase := flag.String("mongodb-database", "engram", "MongoDB database name (mongodb driver)")

	// Project context flags
	repoFlag := flag.String("repo", "", "Project identity (auto-detected from git remote if not set)")

	// Embedder flags
	ollamaURL := flag.String("ollama-url", "http://ollama:11434", "Ollama API URL")
	embeddingModel := flag.String("embedding-model", "nomic-embed-text", "Ollama embedding model")

	// CLI mode flags
	listFlag := flag.Bool("list", false, "List recent memories (CLI mode)")
	limitFlag := flag.Int("limit", 5, "Limit for list operation")
	versionFlag := flag.Bool("version", false, "Print version and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("ec-server %s\n", version)
		return
	}

	ctx := context.Background()

	// Expand paths
	expandedDBPath := expandPath(*dbPath)

	// Build storage config
	cfg := storage.Config{
		Driver:          *storageDriver,
		SQLitePath:      expandedDBPath,
		PostgresDSN:     *postgresDSN,
		MongoDBURI:      *mongoURI,
		MongoDBDatabase: *mongoDatabase,
	}

	// Ensure parent directory exists for SQLite
	if cfg.Driver == "sqlite" {
		if err := os.MkdirAll(filepath.Dir(expandedDBPath), 0755); err != nil {
			log.Fatalf("Failed to create database directory: %v", err)
		}
	}

	// CLI mode - list memories
	if *listFlag {
		if err := runList(ctx, cfg, *limitFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Auto-detect project identity if not provided
	repo := *repoFlag
	if repo == "" {
		repo = gitinfo.GetProjectID()
		log.Printf("Auto-detected project: %s", repo)
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
		Version: version,
	}, nil)

	// Register tools with project context
	tools.RegisterWithRepo(server, svc, repo)

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

	memories, err := store.List(ctx, types.ListOpts{Limit: limit})
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
