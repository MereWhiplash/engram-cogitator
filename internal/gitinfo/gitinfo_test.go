// internal/gitinfo/gitinfo_test.go
package gitinfo_test

import (
	"testing"

	"github.com/MereWhiplash/engram-cogitator/internal/gitinfo"
)

func TestNormalizeRemoteURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"git@github.com:myorg/myrepo.git", "myorg/myrepo"},
		{"https://github.com/myorg/myrepo.git", "myorg/myrepo"},
		{"https://github.com/myorg/myrepo", "myorg/myrepo"},
		{"git@gitlab.com:team/project.git", "team/project"},
		{"ssh://git@github.com/myorg/myrepo.git", "myorg/myrepo"},
		{"https://user:pass@github.com/myorg/myrepo.git", "myorg/myrepo"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := gitinfo.NormalizeRemoteURL(tc.input)
			if result != tc.expected {
				t.Errorf("NormalizeRemoteURL(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestGet(t *testing.T) {
	// Test that Get() doesn't panic and returns a non-nil Info
	// The actual values depend on the git config of the test environment
	info := gitinfo.Get()
	if info == nil {
		t.Fatal("Get() returned nil")
	}

	// Verify the struct fields are accessible (may be empty if not in a git repo)
	_ = info.AuthorName
	_ = info.AuthorEmail
	_ = info.Repo
}
