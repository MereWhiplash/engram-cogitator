// cmd/api/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/MereWhiplash/engram-cogitator/internal/api"
	"github.com/MereWhiplash/engram-cogitator/internal/embedder"
	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
)

func main() {
	// Server flags
	addr := flag.String("addr", ":8080", "Server address")

	// Storage flags
	storageDriver := flag.String("storage-driver", "postgres", "Storage driver: postgres, mongodb")
	postgresDSN := flag.String("postgres-dsn", "", "PostgreSQL connection string")
	mongoURI := flag.String("mongodb-uri", "", "MongoDB connection URI")
	mongoDatabase := flag.String("mongodb-database", "engram", "MongoDB database name")

	// Embedder flags
	ollamaURL := flag.String("ollama-url", "http://localhost:11434", "Ollama API URL")
	embeddingModel := flag.String("embedding-model", "nomic-embed-text", "Ollama embedding model")

	// Migrate flag
	migrateOnly := flag.Bool("migrate", false, "Run migrations and exit")

	// Rate limiting flags
	rateLimit := flag.Int("rate-limit", 100, "Requests per minute per IP (0 to disable)")

	// CORS flags
	corsOrigins := flag.String("cors-origins", "", "Comma-separated list of allowed CORS origins (empty to disable)")

	flag.Parse()

	ctx := context.Background()

	// Build storage config (no sqlite for API server - team mode only)
	cfg := storage.Config{
		Driver:          *storageDriver,
		PostgresDSN:     *postgresDSN,
		MongoDBURI:      *mongoURI,
		MongoDBDatabase: *mongoDatabase,
	}

	// Validate config
	if cfg.Driver == "sqlite" {
		log.Fatal("SQLite not supported for API server - use postgres or mongodb")
	}

	// Initialize storage
	store, err := storage.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// If migrate-only, exit now
	if *migrateOnly {
		log.Println("Migrations complete")
		return
	}

	// Initialize embedder
	emb := embedder.NewOllama(*ollamaURL, *embeddingModel)

	// Create service
	svc := service.New(store, emb)

	// Create handlers
	handlers := api.NewHandlers(svc)

	// Set health check to verify storage connectivity
	handlers.SetHealthCheck(func() error {
		// Simple connectivity check - list with limit 1
		_, err := store.List(context.Background(), storage.ListOpts{Limit: 1})
		return err
	})

	// Setup router
	r := chi.NewRouter()

	// Core middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(api.RequestID)
	r.Use(api.MaxBodySize)

	// Rate limiting (if enabled)
	if *rateLimit > 0 {
		limiter := api.NewRateLimiter(*rateLimit, time.Minute)
		r.Use(limiter.Middleware)
	}

	// CORS (if enabled)
	if *corsOrigins != "" {
		origins := strings.Split(*corsOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		r.Use(api.CORSMiddleware(origins))
	}

	// Git context extraction
	r.Use(api.GitContext)

	// Routes
	r.Get("/health", handlers.Health)
	r.Route("/v1", func(r chi.Router) {
		r.Post("/memories", handlers.Add)
		r.Get("/memories", handlers.List)
		r.Post("/memories/search", handlers.Search)
		r.Put("/memories/{id}/invalidate", handlers.Invalidate)
	})

	// Create server
	srv := &http.Server{
		Addr:         *addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan bool)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}

		close(done)
	}()

	// Start server
	log.Printf("Starting API server on %s", *addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	<-done
	fmt.Println("Server stopped")
}
