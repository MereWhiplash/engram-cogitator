// internal/api/middleware.go
package api

import (
	"context"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Context keys for git info
type contextKey string

const (
	AuthorNameKey  contextKey = "author_name"
	AuthorEmailKey contextKey = "author_email"
	RepoKey        contextKey = "repo"
	RequestIDKey   contextKey = "request_id"
)

// MaxRequestBodySize is the maximum allowed request body size (1MB)
const MaxRequestBodySize = 1 << 20

// repoPattern validates owner/repo format
var repoPattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`)

// RequestID adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID returns request ID from context
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(RequestIDKey); v != nil {
		return v.(string)
	}
	return ""
}

// MaxBodySize limits the request body size to prevent DoS attacks
func MaxBodySize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)
		next.ServeHTTP(w, r)
	})
}

// GitContext extracts git info from headers and adds to context
func GitContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if name := r.Header.Get("X-EC-Author-Name"); name != "" {
			ctx = context.WithValue(ctx, AuthorNameKey, name)
		}
		if email := r.Header.Get("X-EC-Author-Email"); email != "" {
			ctx = context.WithValue(ctx, AuthorEmailKey, email)
		}
		if repo := r.Header.Get("X-EC-Repo"); repo != "" {
			// Validate repo format (owner/repo)
			if repoPattern.MatchString(repo) {
				ctx = context.WithValue(ctx, RepoKey, repo)
			}
			// Invalid format is silently ignored - repo will be empty
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORSMiddleware adds CORS headers for cross-origin requests
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-EC-Author-Name, X-EC-Author-Email, X-EC-Repo, X-Request-ID")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetAuthorName returns author name from context
func GetAuthorName(ctx context.Context) string {
	if v := ctx.Value(AuthorNameKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetAuthorEmail returns author email from context
func GetAuthorEmail(ctx context.Context) string {
	if v := ctx.Value(AuthorEmailKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetRepo returns repo from context
func GetRepo(ctx context.Context) string {
	if v := ctx.Value(RepoKey); v != nil {
		return v.(string)
	}
	return ""
}

// RateLimiter provides simple per-IP rate limiting
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // requests per window
	window   time.Duration // time window
	stop     chan struct{}
	stopOnce sync.Once
}

type visitor struct {
	count    int
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
		stop:     make(chan struct{}),
	}
	// Cleanup old entries periodically
	go rl.cleanup()
	return rl
}

// Stop stops the cleanup goroutine. Safe to call multiple times.
func (rl *RateLimiter) Stop() {
	rl.stopOnce.Do(func() { close(rl.stop) })
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for {
		select {
		case <-rl.stop:
			return
		case <-ticker.C:
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > rl.window {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
		return true
	}

	// Reset if window has passed
	if time.Since(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = time.Now()
		return true
	}

	// Check rate limit
	if v.count >= rl.rate {
		return false
	}

	v.count++
	v.lastSeen = time.Now()
	return true
}

// Middleware returns an HTTP middleware for rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		// Try to get real IP from headers (for proxied requests)
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			ip = realIP
		}

		if !rl.Allow(ip) {
			w.Header().Set("Retry-After", "60")
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
