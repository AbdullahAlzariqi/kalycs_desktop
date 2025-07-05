package watcher

import (
	"context"
	"kalycs/internal/classifier"
	"kalycs/internal/logging"
	"os"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher    *fsnotify.Watcher
	ctx        context.Context
	cancel     context.CancelFunc
	classifier *classifier.Classifier
}

func NewWatcher(ctx_main context.Context, watchPath string, c *classifier.Classifier) (*Watcher, error) {
	ctx, cancel := context.WithCancel(ctx_main)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		cancel()
		return nil, err
	}
	err = watcher.Add(watchPath)
	if err != nil {
		cancel()
		watcher.Close()
		return nil, err
	}

	return &Watcher{
		watcher:    watcher,
		ctx:        ctx,
		cancel:     cancel,
		classifier: c,
	}, nil
}

func (w *Watcher) Start() {
	logging.L().Infow("Starting watcher")

	go func() {
		defer w.watcher.Close()
		logging.L().Debug("Watcher goroutine started")
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					logging.L().Warn("Event channel closed")
					return
				}
				logging.L().Infow("fsnotify event", "event", event, "name", event.Name, "op", event.Op)

				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Rename == fsnotify.Rename {
					info, err := os.Stat(event.Name)
					if err != nil {
						if !os.IsNotExist(err) {
							logging.L().Errorw("failed to stat file after create/rename event", "file", event.Name, "error", err)
						}
						continue
					}
					if !info.IsDir() {
						logging.L().Infow("classifying new file", "path", event.Name)
						if err := w.classifier.Classify(w.ctx, event.Name, info); err != nil {
							logging.L().Errorw("failed to classify file", "file", event.Name, "error", err)
						}
					}
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					logging.L().Warn("Error channel closed")
					return
				}
				logging.L().Errorw("fsnotify error", "error", err)
			case <-w.ctx.Done():
				logging.L().Info("Watcher context done")
				return
			}
		}
	}()
}

func (w *Watcher) Stop() {
	logging.L().Info("Stopping watcher")
	w.cancel()
}
