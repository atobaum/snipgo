package app

import (
	"context"
	"fmt"
	"path/filepath"

	"snipgo/internal/config"
	"snipgo/internal/core"
	"snipgo/internal/history"
	"snipgo/internal/tmpl"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	manager    *core.Manager
	varHistory *history.VarHistory
}

// NewApp creates a new App application struct
func NewApp() (*App, error) {
	manager, err := core.NewDefaultManager()
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
	}

	return app, nil
}

// OnStartup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
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
	return a.manager.Save(snippet)
}

// DeleteSnippet deletes a snippet by ID
func (a *App) DeleteSnippet(id string) error {
	return a.manager.Delete(id)
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
	newManager, err := core.NewDefaultManager()
	if err != nil {
		return fmt.Errorf("failed to create new manager: %w", err)
	}

	a.manager = newManager
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
	if err := a.manager.Save(snippet); err != nil {
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

	// Merge with frontmatter (D10: body is source of truth)
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
