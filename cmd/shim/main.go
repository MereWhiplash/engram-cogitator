// cmd/shim/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MereWhiplash/engram-cogitator/internal/client"
	"github.com/MereWhiplash/engram-cogitator/internal/gitinfo"
	"github.com/MereWhiplash/engram-cogitator/internal/shim"
)

// version is set by goreleaser via ldflags
var version = "dev"

func main() {
	apiURL := flag.String("api-url", "", "Central API URL (required)")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("ec-shim %s\n", version)
		return
	}

	// Check for env var if flag not set
	if *apiURL == "" {
		*apiURL = os.Getenv("EC_API_URL")
	}

	if *apiURL == "" {
		log.Fatal("API URL required: use --api-url or EC_API_URL environment variable")
	}

	// Extract git info
	gitInfo := gitinfo.Get()

	if os.Getenv("EC_DEBUG") != "" {
		log.Printf("Git context: author=%s <%s>, repo=%s", gitInfo.AuthorName, gitInfo.AuthorEmail, gitInfo.Repo)
	}

	// Create API client
	apiClient := client.New(*apiURL, gitInfo)

	// Create shim handler
	handler := shim.NewHandler(apiClient)

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "engram-cogitator",
		Version: version,
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
