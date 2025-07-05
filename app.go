package main

import (
	"context"
	"database/sql"
	"io/fs"
	"kalycs/db"
	"kalycs/internal/classifier"
	"kalycs/internal/logging"
	"kalycs/internal/store"
	"kalycs/internal/utils"
	"kalycs/internal/watcher"
	"path/filepath"
)

// App struct
type App struct {
	ctx        context.Context
	watcher    watcher.Watcher
	db         *sql.DB
	store      *store.Store
	classifier *classifier.Classifier
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

	err := db.InitializeDatabase()
	if err != nil {
		logging.L().Fatalw("Failed to initialize database", "error", err)
	}
	a.db = db.GetDB()

	a.store = store.NewStore(a.db)

	a.classifier = classifier.NewClassifier(a.store)
	if err := a.classifier.LoadIncomingProject(a.ctx); err != nil {
		logging.L().Fatalw("Failed to load incoming project", "error", err)
	}
	if err := a.classifier.Reload(a.ctx); err != nil {
		logging.L().Fatalw("Failed to load rules", "error", err)
	}

	downloadsDir, err := utils.GetDownloadsDirectory()
	if err != nil {
		logging.L().Fatalw("Failed to get downloads directory", "error", err)
	}

	w, err := watcher.NewWatcher(ctx, downloadsDir, a.classifier)
	if err != nil {
		logging.L().Fatalw("Failed to create watcher", "error", err)
	}
	a.watcher = *w
	a.watcher.Start()
}

// domReady is called after the front-end has been loaded
// func (a *App) domReady(ctx context.Context) {
// 	logging.L().Info("DOM ready")
// }

func (a *App) shutdown(ctx context.Context) {
	a.ctx = ctx
	logging.L().Info("Application shutdown")
	a.watcher.Stop()
}

// ImportFolder walks a directory, classifying each file.
func (a *App) ImportFolder(ctx context.Context, dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logging.L().Errorw("error accessing path during import", "path", path, "error", err)
			return err
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			logging.L().Errorw("failed to get file info during import", "path", path, "error", err)
			return nil
		}

		logging.L().Infow("importing and classifying file", "path", path)
		if err := a.classifier.Classify(ctx, path, info); err != nil {
			logging.L().Errorw("failed to classify file during import", "path", path, "error", err)
		}

		return nil
	})
}

// ---------------- Project Methods ----------------

func (a *App) ListProjects(ctx context.Context) ([]db.Project, error) {
	return a.store.Project.GetAll(ctx)
}

func (a *App) CreateProject(ctx context.Context, p db.Project) error {
	return a.store.Project.Create(ctx, &p)
}

func (a *App) UpdateProject(ctx context.Context, p db.Project) error {
	return a.store.Project.Update(ctx, &p)
}

func (a *App) DeleteProject(ctx context.Context, id string) error {
	return a.store.Project.Delete(ctx, id)
}

// ---------------- Rule Methods ----------------

func (a *App) ListRules(ctx context.Context, projectID string) ([]db.Rule, error) {
	return a.store.Rule.GetAllByProject(ctx, projectID)
}

func (a *App) CreateRule(ctx context.Context, r db.Rule) error {
	err := a.store.Rule.Create(ctx, &r)
	if err != nil {
		return err
	}
	return a.classifier.Reload(ctx)
}

func (a *App) UpdateRule(ctx context.Context, r db.Rule) error {
	err := a.store.Rule.Update(ctx, &r)
	if err != nil {
		return err
	}
	return a.classifier.Reload(ctx)
}

func (a *App) DeleteRule(ctx context.Context, id string) error {
	err := a.store.Rule.Delete(ctx, id)
	if err != nil {
		return err
	}
	return a.classifier.Reload(ctx)
}
