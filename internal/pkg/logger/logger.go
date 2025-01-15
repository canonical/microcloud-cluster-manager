// Package logger provides a convience function to constructing a logger
// for use. This is required not just for applications but for testing.
// The logger is exported as a global singleton so that we don't have to inject it as a dependency.
package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is the global logger for the application.
var Log *zap.SugaredLogger

// initializes the global logger for the application.
func init() {
	log, err := newLogger()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	Log = log
}

// newLogger constructs a Sugared Logger that writes to stdout and
// provides human readable timestamps.
func newLogger() (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]any{
		"pod": os.Getenv("POD_NAME"),
	}

	log, err := config.Build()
	if err != nil {
		return nil, err
	}

	return log.Sugar(), nil
}

// SetService sets the service name for the logger.
func SetService(service string) {
	Log = Log.Named(service)
}

// Cleanup ensures that the logger is flushed and all messages are written.
func Cleanup() error {
	if Log != nil {
		return Log.Sync()
	}

	return nil
}
