package operator

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/nekrassov01/ignoresync/internal/params"
)

// RepoInfo holds the repository information for the operator.
type RepoInfo struct {
	Name string // repository name, derived from the remote URL
	Hash string // repository hash, derived from the repository name

	path           string              // local path to the repository
	remote         string              // remote name to use for deriving repository name, e.g. "origin"
	user           string              // git user name
	ignorePatterns []gitignore.Pattern // patterns to ignore
	targetPatterns []gitignore.Pattern // patterns to include as targets
}

// newRepoInfo creates a new RepoInfo by opening the git repository at the given path.
func newRepoInfo(path, remote string) (*RepoInfo, error) {
	// Find repository root
	repoPath, repo, err := findRepoRoot(path)
	if err != nil {
		return nil, NewRepositoryError(fmt.Errorf("failed to open git repository: %w", err))
	}

	// Get repository remote URL
	url, err := getRemoteURL(repo, remote)
	if err != nil {
		return nil, NewRepositoryError(fmt.Errorf("failed to get remote URL: %w", err))
	}

	// Get repository name and hash
	name, hash, err := getRepoName(url)
	if err != nil {
		return nil, NewRepositoryError(fmt.Errorf("failed to get repository name: %w", err))
	}

	// Load ignore patterns from .gitignore
	fs := osfs.New(repoPath)
	ignorePatterns, err := gitignore.ReadPatterns(fs, nil)
	if err != nil {
		return nil, NewRepositoryError(fmt.Errorf("failed to read ignore patterns: %w", err))
	}

	// Get default target patterns
	targetPatterns := getPatterns(defaultPatterns)

	// Get git user name
	user := getUser(repo)

	return &RepoInfo{
		Name:           name,
		Hash:           hash,
		path:           repoPath,
		remote:         remote,
		user:           user,
		ignorePatterns: ignorePatterns,
		targetPatterns: targetPatterns,
	}, nil
}

// findRepoRoot walks up from path until a git repository can be opened.
func findRepoRoot(path string) (string, *git.Repository, error) {
	cur := path
	for {
		if repo, err := git.PlainOpen(cur); err == nil {
			return cur, repo, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", nil, fmt.Errorf("no git repository found starting at %s", path)
		}
		cur = parent
	}
}

// getRemoteURL retrieves the remote URL from the git repository at the given path.
func getRemoteURL(repo *git.Repository, name string) (string, error) {
	if repo == nil {
		return "", errors.New("no repo provided")
	}

	// If remote is not specified, use "origin" by default
	if name == "" {
		name = params.DefaultRemoteName
	}

	// Try to get the specified remote first
	if r, err := repo.Remote(name); err == nil {
		if u := getFirstValidURL(r.Config().URLs); u != "" {
			return u, nil
		}
	}

	// If the specified remote is not found or has no valid URL, try all remotes
	if remotes, err := repo.Remotes(); err == nil {
		for _, remote := range remotes {
			if u := getFirstValidURL(remote.Config().URLs); u != "" {
				return u, nil
			}
		}
	}

	// If no valid remote URL is found, return an error
	return "", fmt.Errorf("no valid remote URL found for %q", name)
}

// getRepoName extracts the repository name from the given URL.
// It returns a SHA-256 hash of the service domain and repository path.
func getRepoName(url string) (string, string, error) {
	name := strings.TrimSuffix(url, ".git")
	if i := strings.Index(name, "://"); i >= 0 {
		name = name[i+3:]
	}
	if i := strings.LastIndex(name, "@"); i >= 0 {
		name = name[i+1:]
	}
	name = strings.Replace(name, ":", "/", 1)
	name = strings.Trim(name, "/")
	parts := strings.Split(name, "/")
	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", errors.New("invalid URL")
	}
	hash := sha256.Sum256([]byte(name))
	return name, hex.EncodeToString(hash[:]), nil
}

// getPatterns retrieves the target patterns for the repository.
func getPatterns(patterns []string) []gitignore.Pattern {
	if len(patterns) == 0 {
		return nil
	}
	targetPatterns := make([]gitignore.Pattern, 0, len(patterns))
	for _, pattern := range patterns {
		targetPatterns = append(targetPatterns, gitignore.ParsePattern(pattern, nil))
	}
	return targetPatterns
}

// getUser retrieves the user name from the git repository or environment variables.
func getUser(repo *git.Repository) string {
	if repo != nil {
		// Get user from local git config
		if cfg, err := repo.Config(); err == nil && cfg.User.Name != "" {
			return cfg.User.Name
		}

		// Get user from git config
		scopes := []config.Scope{config.LocalScope, config.GlobalScope, config.SystemScope}
		for _, scope := range scopes {
			if cfg, err := repo.ConfigScoped(scope); err == nil && cfg.User.Name != "" {
				return cfg.User.Name
			}
		}
	}

	// Get user from environment variables
	if name := os.Getenv("GIT_AUTHOR_NAME"); name != "" {
		return name
	}

	return params.DefaultUserName
}

// getFirstValidURL returns the first non-empty URL from the given list.
func getFirstValidURL(urls []string) string {
	for _, url := range urls {
		if url != "" {
			return url
		}
	}
	return ""
}
