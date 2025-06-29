package utils

import (
	"errors"
	"kalycs/internal/logging"
	"os"
	"path/filepath"
	"runtime"
)

// GetDownloadsDirectory returns the default downloads directory for the current user.
// It supports Windows and macOS. For other operating systems, it returns an error.
func GetDownloadsDirectory() (string, error) {
	logging.L().Info("Attempting to get downloads directory...")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logging.L().Errorw("could not get user home directory", "error", err)
		return "", err
	}

	downloadsPath := filepath.Join(homeDir, "Downloads")

	switch runtime.GOOS {
	case "windows", "darwin":
		logging.L().Infow("Downloads directory found", "os", runtime.GOOS, "path", downloadsPath)
		return downloadsPath, nil
	}

	err = errors.New("unsupported operating system: " + runtime.GOOS)
	logging.L().Warnw("Unsupported operating system", "os", runtime.GOOS)
	return "", err
}
