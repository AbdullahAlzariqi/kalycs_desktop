package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

// Init initialises the global zap logger.
// It chooses the configuration based on the APP_ENV environment variable.
// If APP_ENV is set to "development", a human-friendly development configuration is used.
// Otherwise, a production configuration is created.
func Init() {
	if logger != nil {
		// Already initialised
		return
	}

	var l *zap.Logger
	var err error

	if os.Getenv("APP_ENV") == "development" {
		l, err = zap.NewDevelopment()
	} else {
		// Use production config with ISO8601 timestamps for easier reading
		cfg := zap.NewProductionConfig()
		cfg.Encoding = "json"
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		l, err = cfg.Build()
	}

	if err != nil {
		panic("failed to initialise logger: " + err.Error())
	}

	logger = l.Sugar()
}

// L returns the global *zap.SugaredLogger instance.
// If it has not been initialised yet, Init() will be called implicitly.
func L() *zap.SugaredLogger {
	if logger == nil {
		Init()
	}
	return logger
}
