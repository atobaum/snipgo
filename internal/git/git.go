// Package git provides Git integration for SnipGo snippets.
// It wraps git CLI commands to enable version control, sync, and history features.
package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Common errors
var (
	ErrGitNotInstalled = errors.New("git is not installed. Install git: https://git-scm.com/downloads")
	ErrNotARepository  = errors.New("not a git repository. Run: snipgo git init")
	ErrNoRemote        = errors.New("no remote configured. Run: snipgo git remote add origin <url>")
	ErrMergeConflict   = errors.New("merge conflict detected")
	ErrAuthFailed      = errors.New("authentication failed. Configure SSH key or credential helper")
)

// GitConfig holds git-related configuration
type GitConfig struct {
	Enabled               bool   `yaml:"enabled"`
	AutoCommit            bool   `yaml:"auto_commit"`
	AutoPush              bool   `yaml:"auto_push"`
	CommitMessageTemplate string `yaml:"commit_message_template"`
	Remote                string `yaml:"remote"`
	Branch                string `yaml:"branch"`
}

// DefaultGitConfig returns default git configuration
func DefaultGitConfig() *GitConfig {
	return &GitConfig{
		Enabled:               false,
		AutoCommit:            false,
		AutoPush:              false,
		CommitMessageTemplate: "Update: {{.Title}}",
		Remote:                "origin",
		Branch:                "main",
	}
}

// Commit represents a git commit
type Commit struct {
	Hash      string
	ShortHash string
	Author    string
	Date      time.Time
	Message   string
}

// RepoStatus represents the current repository status
type RepoStatus struct {
	IsRepo          bool
	Branch          string
	HasRemote       bool
	RemoteURL       string
	HasChanges      bool
	UntrackedFiles  []string
	ModifiedFiles   []string
	StagedFiles     []string
	UnpushedCommits int
}

// GitManager handles git operations for the snippets directory
type GitManager struct {
	repoPath string
	config   *GitConfig
}

// NewGitManager creates a new GitManager instance
func NewGitManager(repoPath string, config *GitConfig) *GitManager {
	if config == nil {
		config = DefaultGitConfig()
	}
	return &GitManager{
		repoPath: repoPath,
		config:   config,
	}
}

// IsEnabled returns whether git features are enabled
func (g *GitManager) IsEnabled() bool {
	return g.config != nil && g.config.Enabled
}

// IsGitInstalled checks if git CLI is available
func (g *GitManager) IsGitInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// IsGitRepo checks if the repo path is a git repository
func (g *GitManager) IsGitRepo() bool {
	_, err := g.Exec("rev-parse", "--is-inside-work-tree")
	return err == nil
}

// InitRepo initializes a new git repository
func (g *GitManager) InitRepo() error {
	if !g.IsGitInstalled() {
		return ErrGitNotInstalled
	}

	if g.IsGitRepo() {
		return fmt.Errorf("git repository already initialized at %s", g.repoPath)
	}

	_, err := g.Exec("init")
	if err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Create .gitignore if it doesn't exist
	gitignorePath := filepath.Join(g.repoPath, ".gitignore")
	if _, err := exec.Command("test", "-f", gitignorePath).Output(); err != nil {
		// File doesn't exist, create it
		_, _ = g.Exec("config", "core.excludesfile", gitignorePath)
	}

	return nil
}

// CloneRepo clones a remote repository to the repo path
func (g *GitManager) CloneRepo(url string) error {
	if !g.IsGitInstalled() {
		return ErrGitNotInstalled
	}

	// Clone to a temporary location first, then move contents
	cmd := exec.Command("git", "clone", url, g.repoPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if strings.Contains(errMsg, "Authentication failed") ||
			strings.Contains(errMsg, "Permission denied") {
			return ErrAuthFailed
		}
		return fmt.Errorf("failed to clone repository: %w - %s", err, errMsg)
	}

	return nil
}

// Add stages files for commit
func (g *GitManager) Add(files ...string) error {
	if !g.IsGitRepo() {
		return ErrNotARepository
	}

	args := append([]string{"add"}, files...)
	_, err := g.Exec(args...)
	if err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}
	return nil
}

// AddAll stages all changes
func (g *GitManager) AddAll() error {
	return g.Add("-A")
}

// Commit creates a commit with the given message
func (g *GitManager) Commit(message string) error {
	if !g.IsGitRepo() {
		return ErrNotARepository
	}

	_, err := g.Exec("commit", "-m", message)
	if err != nil {
		// Check if there's nothing to commit
		status, _ := g.Exec("status", "--porcelain")
		if strings.TrimSpace(status) == "" {
			return nil // Nothing to commit is not an error
		}
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

// AddAndCommit stages specific files and commits them
func (g *GitManager) AddAndCommit(message string, files ...string) error {
	if err := g.Add(files...); err != nil {
		return err
	}
	return g.Commit(message)
}

// CommitAll stages all changes and commits
func (g *GitManager) CommitAll(message string) error {
	if err := g.AddAll(); err != nil {
		return err
	}
	return g.Commit(message)
}

// Push pushes commits to remote
func (g *GitManager) Push() error {
	if !g.IsGitRepo() {
		return ErrNotARepository
	}

	if !g.HasRemote() {
		return ErrNoRemote
	}

	remote := g.config.Remote
	branch := g.config.Branch

	_, err := g.Exec("push", "-u", remote, branch)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "Authentication failed") ||
			strings.Contains(errStr, "Permission denied") {
			return ErrAuthFailed
		}
		return fmt.Errorf("failed to push: %w", err)
	}
	return nil
}

// Pull pulls changes from remote
func (g *GitManager) Pull() error {
	if !g.IsGitRepo() {
		return ErrNotARepository
	}

	if !g.HasRemote() {
		return ErrNoRemote
	}

	remote := g.config.Remote
	branch := g.config.Branch

	output, err := g.Exec("pull", remote, branch)
	if err != nil {
		if strings.Contains(output, "CONFLICT") {
			return ErrMergeConflict
		}
		errStr := err.Error()
		if strings.Contains(errStr, "Authentication failed") ||
			strings.Contains(errStr, "Permission denied") {
			return ErrAuthFailed
		}
		return fmt.Errorf("failed to pull: %w", err)
	}
	return nil
}

// Sync performs pull followed by push
func (g *GitManager) Sync() error {
	if err := g.Pull(); err != nil {
		return fmt.Errorf("sync failed during pull: %w", err)
	}
	if err := g.Push(); err != nil {
		return fmt.Errorf("sync failed during push: %w", err)
	}
	return nil
}

// HasRemote checks if a remote is configured
func (g *GitManager) HasRemote() bool {
	output, err := g.Exec("remote")
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) != ""
}

// GetRemoteURL returns the URL of the configured remote
func (g *GitManager) GetRemoteURL() string {
	output, err := g.Exec("remote", "get-url", g.config.Remote)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}

// Status returns the current repository status
func (g *GitManager) Status() (*RepoStatus, error) {
	status := &RepoStatus{
		IsRepo: g.IsGitRepo(),
	}

	if !status.IsRepo {
		return status, nil
	}

	// Get current branch
	branch, _ := g.Exec("branch", "--show-current")
	status.Branch = strings.TrimSpace(branch)

	// Check remote
	status.HasRemote = g.HasRemote()
	if status.HasRemote {
		status.RemoteURL = g.GetRemoteURL()
	}

	// Get file status
	output, err := g.Exec("status", "--porcelain")
	if err != nil {
		return status, fmt.Errorf("failed to get status: %w", err)
	}

	status.HasChanges = strings.TrimSpace(output) != ""

	// Parse status output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		statusCode := line[:2]
		filename := strings.TrimSpace(line[3:])

		switch {
		case statusCode == "??":
			status.UntrackedFiles = append(status.UntrackedFiles, filename)
		case statusCode[0] != ' ':
			status.StagedFiles = append(status.StagedFiles, filename)
		case statusCode[1] != ' ':
			status.ModifiedFiles = append(status.ModifiedFiles, filename)
		}
	}

	// Count unpushed commits
	if status.HasRemote {
		unpushed, _ := g.Exec("rev-list", "--count", fmt.Sprintf("%s/%s..HEAD", g.config.Remote, status.Branch))
		fmt.Sscanf(strings.TrimSpace(unpushed), "%d", &status.UnpushedCommits)
	}

	return status, nil
}

// HasUnpushedCommits returns true if there are local commits not pushed to remote
func (g *GitManager) HasUnpushedCommits() bool {
	status, err := g.Status()
	if err != nil {
		return false
	}
	return status.UnpushedCommits > 0
}

// GetFileHistory returns the commit history for a specific file
func (g *GitManager) GetFileHistory(filepath string) ([]Commit, error) {
	if !g.IsGitRepo() {
		return nil, ErrNotARepository
	}

	// Format: hash|short_hash|author|date|message
	output, err := g.Exec("log", "--follow", "--format=%H|%h|%an|%aI|%s", "--", filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file history: %w", err)
	}

	var commits []Commit
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		date, _ := time.Parse(time.RFC3339, parts[3])
		commits = append(commits, Commit{
			Hash:      parts[0],
			ShortHash: parts[1],
			Author:    parts[2],
			Date:      date,
			Message:   parts[4],
		})
	}

	return commits, nil
}

// GetFileAtCommit returns the content of a file at a specific commit
func (g *GitManager) GetFileAtCommit(filepath, commitHash string) ([]byte, error) {
	if !g.IsGitRepo() {
		return nil, ErrNotARepository
	}

	output, err := g.Exec("show", fmt.Sprintf("%s:%s", commitHash, filepath))
	if err != nil {
		return nil, fmt.Errorf("failed to get file at commit %s: %w", commitHash, err)
	}

	return []byte(output), nil
}

// RestoreFile restores a file to a specific commit version
func (g *GitManager) RestoreFile(filepath, commitHash string) error {
	if !g.IsGitRepo() {
		return ErrNotARepository
	}

	_, err := g.Exec("checkout", commitHash, "--", filepath)
	if err != nil {
		return fmt.Errorf("failed to restore file from commit %s: %w", commitHash, err)
	}

	return nil
}

// GetFileDiff returns the diff between current version and a specific commit
func (g *GitManager) GetFileDiff(filepath, commitHash string) (string, error) {
	if !g.IsGitRepo() {
		return "", ErrNotARepository
	}

	output, err := g.Exec("diff", commitHash, "--", filepath)
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}

	return output, nil
}

// Exec executes a git command and returns the output
func (g *GitManager) Exec(args ...string) (string, error) {
	if !g.IsGitInstalled() {
		return "", ErrGitNotInstalled
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Combine stderr with error for more context
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return stdout.String(), fmt.Errorf("%w: %s", err, errMsg)
		}
		return stdout.String(), err
	}

	return stdout.String(), nil
}

// GetRepoPath returns the repository path
func (g *GitManager) GetRepoPath() string {
	return g.repoPath
}

// GetConfig returns the git configuration
func (g *GitManager) GetConfig() *GitConfig {
	return g.config
}
