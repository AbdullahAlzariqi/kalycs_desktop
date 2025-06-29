package main

import (
	"context"
	"kalycs/internal/logging"
	"kalycs/internal/utils"
	"kalycs/internal/watcher"
)

type Watcher interface {
	Start()
	Stop()
}

// App struct
type App struct {
	ctx     context.Context
	watcher Watcher
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	logging.L().Info("Starting up Kalycs")
	a.ctx = ctx

	downloadsDir, err := utils.GetDownloadsDirectory()
	if err != nil {
		logging.L().Fatalw("Failed to get downloads directory", "error", err)
	}

	w, err := watcher.NewWatcher(ctx, downloadsDir)
	if err != nil {
		logging.L().Fatalw("Failed to create watcher", "error", err)
	}
	a.watcher = w
	a.watcher.Start()
}

// domReady is called after the front-end has been loaded
func (a *App) domReady(ctx context.Context) {
	logging.L().Info("DOM ready")
}

func (a *App) shutdown(ctx context.Context) {
	a.ctx = ctx
	logging.L().Info("Application shutdown")
	a.watcher.Stop()
}
