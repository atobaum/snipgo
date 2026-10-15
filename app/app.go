package app

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"snipgo/internal/config"
	"snipgo/internal/core"
	"snipgo/internal/gitsync"
	"snipgo/internal/history"
	"snipgo/internal/tmpl"
	"snipgo/internal/watcher"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	snippetsChangedEvent = "snippets:changed"
	syncStatusEvent      = "sync:status"
	watchDebounce        = 200 * time.Millisecond
	syncTimeout          = 2 * time.Minute
)

// App struct
type App struct {
	ctx        context.Context
	manager    *core.Manager
	varHistory *history.VarHistory

	watcherMu sync.Mutex
	watcher   *watcher.Watcher

	syncMu    sync.Mutex
	scheduler *gitsync.Scheduler

	statusMu   sync.Mutex
	syncStatus gitsync.Status
}

// NewApp creates a new App application struct
func NewApp() (*App, error) {
	manager, err := core.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Initialize variable history
	configPath, err := config.GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}
	historyPath := filepath.Join(filepath.Dir(configPath), "var_history.json")
	varHistory, err := history.NewVarHistory(historyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create var history: %w", err)
	}

	app := &App{
		manager:    manager,
		varHistory: varHistory,
		syncStatus: gitsync.Status{ConflictFiles: []string{}},
	}

	// Load all snippets on startup
	if err := manager.LoadAll(); err != nil {
		return nil, fmt.Errorf("failed to load snippets: %w", err)
	}

	return app, nil
}

// OnStartup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.restartWatcher()
	a.restartSync()
}

// OnShutdown stops watching the data directory.
func (a *App) OnShutdown(_ context.Context) {
	a.stopSync()
	a.watcherMu.Lock()
	defer a.watcherMu.Unlock()
	a.closeWatcher()
}

// restartWatcher (re)starts hot reload for the current manager's data directory.
// The callback keeps the manager it was started with so a later manager swap cannot race with it.
func (a *App) restartWatcher() {
	a.watcherMu.Lock()
	defer a.watcherMu.Unlock()
	a.closeWatcher()

	manager := a.manager
	w, err := watcher.New(manager.GetDataDir(), watchDebounce, func(_ []string) {
		if err := manager.LoadAll(); err != nil {
			slog.Warn("hot reload failed to reload snippets", "error", err)
			return
		}
		runtime.EventsEmit(a.ctx, snippetsChangedEvent)
	})
	if err != nil {
		slog.Warn("hot reload disabled", "dir", manager.GetDataDir(), "error", err)
		return
	}
	a.watcher = w
}

// restartSync rebuilds the scheduler from the current config and data directory.
func (a *App) restartSync() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Warn("sync disabled: failed to load config", "error", err)
		return
	}

	a.syncMu.Lock()
	defer a.syncMu.Unlock()
	if a.scheduler != nil {
		a.scheduler.Stop()
	}

	syncer := gitsync.New(gitsync.ExecRunner{}, gitsync.Options{
		Dir:      a.manager.GetDataDir(),
		Conflict: gitsync.Mode(cfg.Sync.Conflict),
	})
	run := func(ctx context.Context) (gitsync.Result, error) {
		ctx, cancel := context.WithTimeout(ctx, syncTimeout)
		defer cancel()
		return syncer.Sync(ctx)
	}
	a.scheduler = gitsync.NewScheduler(run, time.Duration(cfg.Sync.IntervalMinutes)*time.Minute, a.onSyncEvent)
	if cfg.Sync.Enabled {
		a.scheduler.Start()
	}
}

func (a *App) stopSync() {
	a.syncMu.Lock()
	defer a.syncMu.Unlock()
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
}

// onSyncEvent must not take syncMu: Stop waits for the run that calls it.
func (a *App) onSyncEvent(ev gitsync.Event) {
	status := gitsync.NewStatus(ev)
	a.statusMu.Lock()
	a.syncStatus = status
	a.statusMu.Unlock()
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, syncStatusEvent, status)
	}
}

// SyncNow runs a manual sync and returns the resulting status; errors are reported in the status.
func (a *App) SyncNow() gitsync.Status {
	a.syncMu.Lock()
	scheduler := a.scheduler
	a.syncMu.Unlock()
	if scheduler == nil {
		a.restartSync()
		a.syncMu.Lock()
		scheduler = a.scheduler
		a.syncMu.Unlock()
	}
	if scheduler != nil {
		_, _ = scheduler.RunNow(context.Background())
	}
	return a.GetSyncStatus()
}

func (a *App) GetSyncStatus() gitsync.Status {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	return a.syncStatus
}

func (a *App) GetSyncSettings() (config.SyncConfig, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return config.SyncConfig{}, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg.Sync, nil
}

func (a *App) SaveSyncSettings(s config.SyncConfig) error {
	if err := config.ValidateSync(s); err != nil {
		return err
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg.Sync = s
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	a.restartSync()
	return nil
}

// closeWatcher must be called with watcherMu held.
func (a *App) closeWatcher() {
	if a.watcher == nil {
		return
	}
	if err := a.watcher.Close(); err != nil {
		slog.Warn("failed to close file watcher", "error", err)
	}
	a.watcher = nil
}

func (a *App) markSelfWrite(paths ...string) {
	a.watcherMu.Lock()
	defer a.watcherMu.Unlock()
	if a.watcher == nil {
		return
	}
	for _, p := range paths {
		if p != "" {
			a.watcher.MarkSelfWrite(p)
		}
	}
}

// saveSnippet saves through the manager and marks both the old and new file paths as self-written.
func (a *App) saveSnippet(snippet *core.Snippet) error {
	oldPath, _ := a.manager.FilePath(snippet.ID)
	err := a.manager.Save(snippet)
	newPath, _ := a.manager.FilePath(snippet.ID)
	a.markSelfWrite(oldPath, newPath)
	return err
}

// GetAllSnippets returns all snippets
func (a *App) GetAllSnippets() ([]*core.Snippet, error) {
	return a.manager.GetAll(), nil
}

// GetSnippet returns a snippet by ID
func (a *App) GetSnippet(id string) (*core.Snippet, error) {
	return a.manager.GetByID(id)
}

// SaveSnippet saves a snippet
func (a *App) SaveSnippet(snippet *core.Snippet) error {
	return a.saveSnippet(snippet)
}

// DeleteSnippet deletes a snippet by ID
func (a *App) DeleteSnippet(id string) error {
	path, _ := a.manager.FilePath(id)
	err := a.manager.Delete(id)
	a.markSelfWrite(path)
	return err
}

// ConfirmDialog shows a native Yes/No dialog because window.confirm is a no-op in the macOS webview.
func (a *App) ConfirmDialog(title, message string) (bool, error) {
	result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         title,
		Message:       message,
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "Yes",
		CancelButton:  "No",
	})
	if err != nil {
		return false, err
	}
	return result == "Yes", nil
}

// AlertDialog shows a native info dialog because window.alert is a no-op in the macOS webview.
func (a *App) AlertDialog(title, message string) error {
	_, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
	})
	return err
}

// SearchSnippets searches snippets by query
func (a *App) SearchSnippets(query string) ([]*core.Snippet, error) {
	results := a.manager.Search(query)
	snippets := make([]*core.Snippet, len(results))
	for i, result := range results {
		snippets[i] = result.Snippet
	}
	return snippets, nil
}

// CopyToClipboard copies text to clipboard
func (a *App) CopyToClipboard(text string) error {
	return runtime.ClipboardSetText(a.ctx, text)
}

// ReloadSnippets reloads all snippets from disk
func (a *App) ReloadSnippets() error {
	return a.manager.LoadAll()
}

// GetAllTags returns all unique tags
func (a *App) GetAllTags() ([]string, error) {
	return a.manager.GetAllTags(), nil
}

// GetConfigPath returns the current config file path
func (a *App) GetConfigPath() (string, error) {
	return config.GetConfigPath()
}

// GetDataDirectory returns the current data directory
func (a *App) GetDataDirectory() (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", err
	}
	return cfg.DataDirectory, nil
}

// SetDataDirectory updates the data directory and reloads snippets
func (a *App) SetDataDirectory(path string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.DataDirectory = path
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Recreate manager with new config
	newManager, err := core.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create new manager: %w", err)
	}

	if err := newManager.LoadAll(); err != nil {
		return fmt.Errorf("failed to load snippets: %w", err)
	}

	a.manager = newManager
	a.restartWatcher()
	a.restartSync()
	return nil
}

// BrowseForDirectory opens a directory picker dialog
func (a *App) BrowseForDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Snippets Directory",
	})
}

// CreateSnippet creates a new snippet with the given title
func (a *App) CreateSnippet(title string) (*core.Snippet, error) {
	snippet := core.NewSnippet(title)
	if err := a.saveSnippet(snippet); err != nil {
		return nil, err
	}
	return snippet, nil
}

// ExtractVariables returns variables found in a snippet's body, enriched with
// frontmatter metadata. Follows D10 merge rules: body is source of truth.
func (a *App) ExtractVariables(snippetID string) ([]*tmpl.Variable, error) {
	snippet, err := a.manager.GetByID(snippetID)
	if err != nil {
		return nil, err
	}

	bodyVarNames := tmpl.ExtractVariables(snippet.Body)
	if len(bodyVarNames) == 0 {
		return []*tmpl.Variable{}, nil
	}

	return tmpl.MergeWithMetadata(bodyVarNames, snippet.Variables), nil
}

// ExpandSnippet expands a snippet's body with the given variable values.
func (a *App) ExpandSnippet(snippetID string, values map[string]string) (string, error) {
	snippet, err := a.manager.GetByID(snippetID)
	if err != nil {
		return "", err
	}

	result, err := tmpl.Expand(snippet.Body, values)
	if err != nil {
		return "", err
	}
	return result.Expanded, nil
}

// GetVariableHistory returns recent values for a variable name.
func (a *App) GetVariableHistory(varName string) ([]string, error) {
	if a.varHistory == nil {
		return []string{}, nil
	}
	return a.varHistory.Get(varName), nil
}

// SaveVariableHistory saves variable values to history.
func (a *App) SaveVariableHistory(values map[string]string) error {
	if a.varHistory == nil {
		return nil
	}
	for name, value := range values {
		if err := a.varHistory.Add(name, value); err != nil {
			return err
		}
	}
	return nil
}
