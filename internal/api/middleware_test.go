// internal/api/middleware_test.go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
