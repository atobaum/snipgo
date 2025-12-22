package app

import (
	"context"
	"fmt"

	"snipgo/internal/core"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	manager *core.Manager
}

// NewApp creates a new App application struct
func NewApp() (*App, error) {
	manager, err := core.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	app := &App{
		manager: manager,
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
