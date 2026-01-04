package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MereWhiplash/engram-cogitator/internal/db"
	"github.com/MereWhiplash/engram-cogitator/internal/embed"
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

	// Initialize database
	database, err := db.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize embedding client
	embedder := embed.New(*ollamaURL, *embeddingModel)

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "engram-cogitator",
		Version: "1.0.0",
	}, nil)

	// Register tools
	tools.Register(server, database, embedder)

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

// runList handles CLI mode for listing memories
func runList(dbPath string, limit int) error {
	database, err := db.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	memories, err := database.List(limit, "", "", false)
	if err != nil {
		return fmt.Errorf("failed to list memories: %w", err)
	}

	if len(memories) == 0 {
		return nil // Silent exit if no memories
	}

	for _, m := range memories {
		fmt.Printf("[%s/%s] %s\n", m.Type, m.Area, m.Content)
	}
	return nil
}
