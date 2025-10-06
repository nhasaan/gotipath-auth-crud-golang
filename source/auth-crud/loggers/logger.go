package loggers

import (
	"encoding/json"
	"fmt"
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
	logger   *log.Logger
	initOnce sync.Once
)

func initLogger() {
	outputMode := os.Getenv("LOG_OUTPUT")
	if outputMode != "file" {
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
			writer = os.Stdout
		}
	}

	logger = log.New(writer, "", 0)
}

// L returns the initialized *log.Logger singleton.
func L() *log.Logger {
	initOnce.Do(initLogger)
	return logger
}

// Log writes a single line JSON log with provided fields.
func Log(fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	// ensure timestamp exists
	if _, ok := fields["ts"]; !ok {
		fields["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	}
	b, err := json.Marshal(fields)
	if err != nil {
		// fallback
		L().Print(fmt.Sprintf(`{"ts":"%s","level":"error","msg":"failed to marshal log","err":"%v"}`, time.Now().UTC().Format(time.RFC3339Nano), err))
		return
	}
	L().Print(string(b))
}

// Info logs a message at info level in JSON form.
func Info(v ...any) {
	Log(map[string]interface{}{"level": "info", "msg": fmt.Sprint(v...)})
}

// Error logs an error message in JSON form.
func Error(v ...any) {
	Log(map[string]interface{}{"level": "error", "msg": fmt.Sprint(v...)})
}
