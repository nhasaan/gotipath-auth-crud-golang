package loggers

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Environment variables:
// LOG_OUTPUT: "stdout" (default) or "file"
// LOG_FILE_PATH: when LOG_OUTPUT=file, path to log file (default: /app/logs/app.log)

var (
	logger     *log.Logger
	initOnce   sync.Once
	logOutputs = map[string]bool{"stdout": true, "file": true}
)

func initLogger() {
	outputMode := os.Getenv("LOG_OUTPUT")
	if !logOutputs[outputMode] {
		outputMode = "stdout"
	}

	var writer io.Writer = os.Stdout
	if outputMode == "file" {
		path := os.Getenv("LOG_FILE_PATH")
		if path == "" {
			path = "/app/logs/app.log"
		}
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			writer = f
		} else {
			// Fallback to stdout if we cannot open the file
			writer = os.Stdout
		}
	}

	logger = log.New(writer, "", log.LstdFlags|log.LUTC)
}

// L returns the initialized *log.Logger singleton.
func L() *log.Logger {
	initOnce.Do(initLogger)
	return logger
}

// Info logs an informational message.
func Info(v ...any) {
	L().Println(v...)
}

// Error logs an error message with a consistent prefix.
func Error(v ...any) {
	L().Println(append([]any{"ERROR:"}, v...)...)
}

// WithTime allows logging with a custom timestamp (UTC) when needed.
func WithTime(t time.Time, v ...any) {
	// temporarily create a logger with the provided timestamp prefix
	prefix := t.UTC().Format("2006/01/02 15:04:05 ")
	l := log.New(L().Writer(), prefix, 0)
	l.Println(v...)
}
