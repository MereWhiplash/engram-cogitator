// internal/gitinfo/gitinfo.go
package gitinfo

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Info holds git configuration info
type Info struct {
	AuthorName  string
	AuthorEmail string
	Repo        string
}

// Get extracts git info from the current directory.
// Returns partial info if some git commands fail (e.g., not in a git repo).
func Get() *Info {
	info := &Info{}

	// Get author name
	if out, err := exec.Command("git", "config", "user.name").Output(); err == nil {
		info.AuthorName = strings.TrimSpace(string(out))
	}

	// Get author email
	if out, err := exec.Command("git", "config", "user.email").Output(); err == nil {
		info.AuthorEmail = strings.TrimSpace(string(out))
	}

	// Get remote URL
	if out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output(); err == nil {
		info.Repo = NormalizeRemoteURL(strings.TrimSpace(string(out)))
	}

	return info
}

// GetProjectID returns a unique identifier for the current project.
// Priority:
// 1. Git remote origin → normalized to "org/repo"
// 2. Working directory absolute path (for non-git projects)
func GetProjectID() string {
	// Try git remote first
	if out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output(); err == nil {
		repo := NormalizeRemoteURL(strings.TrimSpace(string(out)))
		if repo != "" {
			return repo
		}
	}

	// Fall back to working directory
	if wd, err := os.Getwd(); err == nil {
		// Clean and use absolute path
		return filepath.Clean(wd)
	}

	return "unknown"
}

// NormalizeRemoteURL converts various git remote URL formats to "org/repo"
func NormalizeRemoteURL(url string) string {
	url = strings.TrimSpace(url)

	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH format: git@github.com:org/repo
	if strings.HasPrefix(url, "git@") {
		// git@github.com:org/repo -> org/repo
		re := regexp.MustCompile(`git@[^:]+:(.+)`)
		if matches := re.FindStringSubmatch(url); len(matches) > 1 {
			return matches[1]
		}
	}

	// Handle ssh:// format: ssh://git@github.com/org/repo
	if strings.HasPrefix(url, "ssh://") {
		url = strings.TrimPrefix(url, "ssh://")
		// Remove user@host part
		if idx := strings.Index(url, "/"); idx != -1 {
			url = url[idx+1:]
		}
		return url
	}

	// Handle HTTPS format: https://github.com/org/repo
	// Also handles https://user:pass@github.com/org/repo
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		// Remove protocol
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "http://")

		// Remove user:pass@ if present
		if idx := strings.Index(url, "@"); idx != -1 {
			url = url[idx+1:]
		}

		// Remove host
		if idx := strings.Index(url, "/"); idx != -1 {
			url = url[idx+1:]
		}
		return url
	}

	return url
}
