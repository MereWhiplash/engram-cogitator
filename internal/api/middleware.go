// internal/api/middleware.go
package api

import (
	"context"
	"net/http"
)

// Context keys for git info
type contextKey string

const (
	AuthorNameKey  contextKey = "author_name"
	AuthorEmailKey contextKey = "author_email"
	RepoKey        contextKey = "repo"
)

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
			ctx = context.WithValue(ctx, RepoKey, repo)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
