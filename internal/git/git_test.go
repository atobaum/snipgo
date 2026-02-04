package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewGitManager(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		gm := NewGitManager("/tmp/test", nil)
		if gm == nil {
			t.Fatal("expected non-nil GitManager")
		}
		if gm.config == nil {
			t.Fatal("expected non-nil config")
		}
		if gm.config.Remote != "origin" {
			t.Errorf("expected default remote 'origin', got %s", gm.config.Remote)
		}
		if gm.config.Branch != "main" {
			t.Errorf("expected default branch 'main', got %s", gm.config.Branch)
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := &GitConfig{
			Enabled:    true,
			AutoCommit: true,
			Remote:     "upstream",
			Branch:     "develop",
		}
		gm := NewGitManager("/tmp/test", cfg)
		if gm.config.Remote != "upstream" {
			t.Errorf("expected remote 'upstream', got %s", gm.config.Remote)
		}
		if gm.config.Branch != "develop" {
			t.Errorf("expected branch 'develop', got %s", gm.config.Branch)
		}
	})
}

func TestIsEnabled(t *testing.T) {
	t.Run("disabled by default", func(t *testing.T) {
		gm := NewGitManager("/tmp/test", nil)
		if gm.IsEnabled() {
			t.Error("expected IsEnabled to be false by default")
		}
	})

	t.Run("enabled when config says so", func(t *testing.T) {
		cfg := &GitConfig{Enabled: true}
		gm := NewGitManager("/tmp/test", cfg)
		if !gm.IsEnabled() {
			t.Error("expected IsEnabled to be true")
		}
	})
}

func TestIsGitInstalled(t *testing.T) {
	gm := NewGitManager("/tmp/test", nil)
	// This test assumes git is installed on the test machine
	if !gm.IsGitInstalled() {
		t.Skip("git is not installed, skipping test")
	}
}

func TestIsGitRepo(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "snipgo-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gm := NewGitManager(tmpDir, nil)

	if !gm.IsGitInstalled() {
		t.Skip("git is not installed, skipping test")
	}

	t.Run("returns false for non-repo", func(t *testing.T) {
		if gm.IsGitRepo() {
			t.Error("expected IsGitRepo to be false for non-repo directory")
		}
	})

	t.Run("returns true after init", func(t *testing.T) {
		if err := gm.InitRepo(); err != nil {
			t.Fatalf("failed to init repo: %v", err)
		}
		if !gm.IsGitRepo() {
			t.Error("expected IsGitRepo to be true after init")
		}
	})
}

func TestInitRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gm := NewGitManager(tmpDir, nil)

	if !gm.IsGitInstalled() {
		t.Skip("git is not installed, skipping test")
	}

	t.Run("initializes new repo", func(t *testing.T) {
		if err := gm.InitRepo(); err != nil {
			t.Fatalf("failed to init repo: %v", err)
		}

		// Check .git directory exists
		gitDir := filepath.Join(tmpDir, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			t.Error("expected .git directory to exist")
		}
	})

	t.Run("returns error if already initialized", func(t *testing.T) {
		err := gm.InitRepo()
		if err == nil {
			t.Error("expected error when initializing already initialized repo")
		}
	})
}

func TestAddAndCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gm := NewGitManager(tmpDir, nil)

	if !gm.IsGitInstalled() {
		t.Skip("git is not installed, skipping test")
	}

	// Initialize repo
	if err := gm.InitRepo(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	// Configure git user for commits and disable GPG signing
	gm.Exec("config", "user.email", "test@example.com")
	gm.Exec("config", "user.name", "Test User")
	gm.Exec("config", "commit.gpgsign", "false")
	gm.Exec("config", "commit.gpgsign", "false")

	t.Run("commits new file", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tmpDir, "test.md")
		if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Add and commit
		if err := gm.AddAndCommit("Add test file", "test.md"); err != nil {
			t.Fatalf("failed to add and commit: %v", err)
		}

		// Verify commit was created
		output, err := gm.Exec("log", "--oneline", "-1")
		if err != nil {
			t.Fatalf("failed to get log: %v", err)
		}
		if output == "" {
			t.Error("expected commit in log")
		}
	})
}

func TestCommitAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gm := NewGitManager(tmpDir, nil)

	if !gm.IsGitInstalled() {
		t.Skip("git is not installed, skipping test")
	}

	// Initialize repo and configure user
	if err := gm.InitRepo(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	gm.Exec("config", "user.email", "test@example.com")
	gm.Exec("config", "user.name", "Test User")
	gm.Exec("config", "commit.gpgsign", "false")
	gm.Exec("config", "commit.gpgsign", "false")

	t.Run("commits all changes", func(t *testing.T) {
		// Create multiple test files
		for _, name := range []string{"file1.md", "file2.md", "file3.md"} {
			testFile := filepath.Join(tmpDir, name)
			if err := os.WriteFile(testFile, []byte("# "+name), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}
		}

		// Commit all
		if err := gm.CommitAll("Add all files"); err != nil {
			t.Fatalf("failed to commit all: %v", err)
		}

		// Verify files are tracked
		output, err := gm.Exec("ls-files")
		if err != nil {
			t.Fatalf("failed to list files: %v", err)
		}
		for _, name := range []string{"file1.md", "file2.md", "file3.md"} {
			if !contains(output, name) {
				t.Errorf("expected %s to be tracked", name)
			}
		}
	})
}

func TestStatus(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gm := NewGitManager(tmpDir, nil)

	if !gm.IsGitInstalled() {
		t.Skip("git is not installed, skipping test")
	}

	t.Run("returns not a repo status", func(t *testing.T) {
		status, err := gm.Status()
		if err != nil {
			t.Fatalf("failed to get status: %v", err)
		}
		if status.IsRepo {
			t.Error("expected IsRepo to be false")
		}
	})

	// Initialize repo
	if err := gm.InitRepo(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	gm.Exec("config", "user.email", "test@example.com")
	gm.Exec("config", "user.name", "Test User")
	gm.Exec("config", "commit.gpgsign", "false")

	t.Run("detects untracked files", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "untracked.md")
		if err := os.WriteFile(testFile, []byte("# Untracked"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		status, err := gm.Status()
		if err != nil {
			t.Fatalf("failed to get status: %v", err)
		}

		if !status.HasChanges {
			t.Error("expected HasChanges to be true")
		}
		if len(status.UntrackedFiles) == 0 {
			t.Error("expected untracked files")
		}
	})

	t.Run("detects modified files", func(t *testing.T) {
		// First commit the file
		gm.CommitAll("Initial commit")

		// Modify the file
		testFile := filepath.Join(tmpDir, "untracked.md")
		if err := os.WriteFile(testFile, []byte("# Modified"), 0644); err != nil {
			t.Fatalf("failed to modify test file: %v", err)
		}

		status, err := gm.Status()
		if err != nil {
			t.Fatalf("failed to get status: %v", err)
		}

		if !status.HasChanges {
			t.Error("expected HasChanges to be true")
		}
		if len(status.ModifiedFiles) == 0 {
			t.Error("expected modified files")
		}
	})
}

func TestGetFileHistory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gm := NewGitManager(tmpDir, nil)

	if !gm.IsGitInstalled() {
		t.Skip("git is not installed, skipping test")
	}

	// Initialize repo and configure user
	if err := gm.InitRepo(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	gm.Exec("config", "user.email", "test@example.com")
	gm.Exec("config", "user.name", "Test User")
	gm.Exec("config", "commit.gpgsign", "false")

	testFile := filepath.Join(tmpDir, "history-test.md")

	// Create multiple commits
	for i := 1; i <= 3; i++ {
		content := []byte("# Version " + string(rune('0'+i)))
		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		if err := gm.AddAndCommit("Version "+string(rune('0'+i)), "history-test.md"); err != nil {
			t.Fatalf("failed to commit: %v", err)
		}
	}

	t.Run("returns commit history", func(t *testing.T) {
		commits, err := gm.GetFileHistory("history-test.md")
		if err != nil {
			t.Fatalf("failed to get history: %v", err)
		}

		if len(commits) != 3 {
			t.Errorf("expected 3 commits, got %d", len(commits))
		}

		// Most recent commit should be first
		if commits[0].Message != "Version 3" {
			t.Errorf("expected most recent commit first, got %s", commits[0].Message)
		}
	})
}

func TestGetFileAtCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gm := NewGitManager(tmpDir, nil)

	if !gm.IsGitInstalled() {
		t.Skip("git is not installed, skipping test")
	}

	// Initialize repo and configure user
	if err := gm.InitRepo(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	gm.Exec("config", "user.email", "test@example.com")
	gm.Exec("config", "user.name", "Test User")
	gm.Exec("config", "commit.gpgsign", "false")

	testFile := filepath.Join(tmpDir, "content-test.md")

	// First version
	if err := os.WriteFile(testFile, []byte("Original content"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := gm.AddAndCommit("Original", "content-test.md"); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Get the commit hash
	commits, _ := gm.GetFileHistory("content-test.md")
	originalHash := commits[0].Hash

	// Modify file
	if err := os.WriteFile(testFile, []byte("Modified content"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := gm.AddAndCommit("Modified", "content-test.md"); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	t.Run("retrieves content at specific commit", func(t *testing.T) {
		content, err := gm.GetFileAtCommit("content-test.md", originalHash)
		if err != nil {
			t.Fatalf("failed to get file at commit: %v", err)
		}

		if string(content) != "Original content" {
			t.Errorf("expected 'Original content', got '%s'", string(content))
		}
	})
}

func TestHasRemote(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gm := NewGitManager(tmpDir, nil)

	if !gm.IsGitInstalled() {
		t.Skip("git is not installed, skipping test")
	}

	// Initialize repo
	if err := gm.InitRepo(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	t.Run("returns false when no remote", func(t *testing.T) {
		if gm.HasRemote() {
			t.Error("expected HasRemote to be false")
		}
	})

	t.Run("returns true after adding remote", func(t *testing.T) {
		gm.Exec("remote", "add", "origin", "https://github.com/example/repo.git")
		if !gm.HasRemote() {
			t.Error("expected HasRemote to be true")
		}
	})
}

func TestExec(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo-git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	gm := NewGitManager(tmpDir, nil)

	if !gm.IsGitInstalled() {
		t.Skip("git is not installed, skipping test")
	}

	t.Run("executes git command", func(t *testing.T) {
		output, err := gm.Exec("version")
		if err != nil {
			t.Fatalf("failed to execute git version: %v", err)
		}
		if !contains(output, "git version") {
			t.Error("expected output to contain 'git version'")
		}
	})

	t.Run("returns error for invalid command", func(t *testing.T) {
		_, err := gm.Exec("invalid-command-xyz")
		if err == nil {
			t.Error("expected error for invalid command")
		}
	})
}

func TestDefaultGitConfig(t *testing.T) {
	cfg := DefaultGitConfig()

	if cfg.Enabled {
		t.Error("expected Enabled to be false by default")
	}
	if cfg.AutoCommit {
		t.Error("expected AutoCommit to be false by default")
	}
	if cfg.AutoPush {
		t.Error("expected AutoPush to be false by default")
	}
	if cfg.Remote != "origin" {
		t.Errorf("expected Remote to be 'origin', got %s", cfg.Remote)
	}
	if cfg.Branch != "main" {
		t.Errorf("expected Branch to be 'main', got %s", cfg.Branch)
	}
	if cfg.CommitMessageTemplate != "Update: {{.Title}}" {
		t.Errorf("unexpected commit message template: %s", cfg.CommitMessageTemplate)
	}
}

func TestErrors(t *testing.T) {
	t.Run("ErrNotARepository for non-repo operations", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "snipgo-git-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		gm := NewGitManager(tmpDir, nil)

		if !gm.IsGitInstalled() {
			t.Skip("git is not installed, skipping test")
		}

		err = gm.Add("test.md")
		if err == nil {
			t.Error("expected error for Add on non-repo")
		}

		err = gm.Commit("test")
		if err == nil {
			t.Error("expected error for Commit on non-repo")
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
