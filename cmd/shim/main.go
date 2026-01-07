// cmd/shim/main.go
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MereWhiplash/engram-cogitator/internal/client"
	"github.com/MereWhiplash/engram-cogitator/internal/gitinfo"
	"github.com/MereWhiplash/engram-cogitator/internal/shim"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	apiURL := flag.String("api-url", "", "Central API URL (required)")
	flag.Parse()

	// Check for env var if flag not set
	if *apiURL == "" {
		*apiURL = os.Getenv("EC_API_URL")
	}

	if *apiURL == "" {
		log.Fatal("API URL required: use --api-url or EC_API_URL environment variable")
	}

	// Extract git info
	gitInfo, err := gitinfo.Get()
	if err != nil {
		log.Printf("Warning: failed to get git info: %v", err)
		gitInfo = &gitinfo.Info{}
	}

	log.Printf("Git context: author=%s <%s>, repo=%s", gitInfo.AuthorName, gitInfo.AuthorEmail, gitInfo.Repo)

	// Create API client
	apiClient := client.New(*apiURL, gitInfo)

	// Create shim handler
	handler := shim.NewHandler(apiClient)

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "engram-cogitator",
		Version: "1.0.0",
	}, nil)

	// Register tools
	shim.Register(server, handler)

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
	log.Println("Starting Engram Cogitator shim...")
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
