// internal/api/middleware_test.go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MereWhiplash/engram-cogitator/internal/api"
)

func TestGitContext(t *testing.T) {
	var capturedName, capturedEmail, capturedRepo string

	handler := api.GitContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedName = api.GetAuthorName(r.Context())
		capturedEmail = api.GetAuthorEmail(r.Context())
		capturedRepo = api.GetRepo(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-EC-Author-Name", "Alice")
	req.Header.Set("X-EC-Author-Email", "alice@example.com")
	req.Header.Set("X-EC-Repo", "myorg/myrepo")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if capturedName != "Alice" {
		t.Errorf("expected name 'Alice', got %q", capturedName)
	}
	if capturedEmail != "alice@example.com" {
		t.Errorf("expected email 'alice@example.com', got %q", capturedEmail)
	}
	if capturedRepo != "myorg/myrepo" {
		t.Errorf("expected repo 'myorg/myrepo', got %q", capturedRepo)
	}
}

func TestRateLimiter(t *testing.T) {
	rl := api.NewRateLimiter(3, 100*time.Millisecond)
	defer rl.Stop()

	// First 3 should be allowed
	for i := 0; i < 3; i++ {
		if !rl.Allow("192.168.1.1") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 4th should be denied
	if rl.Allow("192.168.1.1") {
		t.Error("4th request should be denied")
	}

	// Different IP should be allowed
	if !rl.Allow("192.168.1.2") {
		t.Error("different IP should be allowed")
	}

	// Wait for window to pass, then should be allowed again
	time.Sleep(150 * time.Millisecond)
	if !rl.Allow("192.168.1.1") {
		t.Error("request after window should be allowed")
	}
}

func TestRateLimiterStop(t *testing.T) {
	rl := api.NewRateLimiter(10, 10*time.Millisecond)

	// Allow a request to ensure cleanup goroutine is running
	rl.Allow("test")

	// Stop should not panic or block
	rl.Stop()

	// Calling Stop again should be safe (sync.Once prevents double-close panic)
	rl.Stop()
	rl.Stop() // Call multiple times to ensure it's truly safe
}
