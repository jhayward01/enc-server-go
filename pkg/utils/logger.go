package utils

import (
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const defaultDir = "/tmp/enc-server-go-logs"

// InitLogger configures structured JSON logging via log/slog.
// It returns the log file handle (to be closed by caller) or an error.
func InitLogger(serviceName string) (*os.File, error) {
	// 1. Determine log directory override.
	logDir := defaultDir
	if val, ok := os.LookupEnv("ENC_SERVER_GO_LOG_DIR"); ok {
		logDir = val
	}

	// 2. Ensure log directory exists.
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	// 3. Create log file with RFC3339 timestamp.
	timestamp := time.Now().UTC().Format("20060102-150405")
	fileName := serviceName + "." + timestamp + ".log"
	logPath := filepath.Join(logDir, fileName)

	logFileFlags := os.O_CREATE | os.O_APPEND | os.O_RDWR
	logFilePerm := fs.FileMode(0666)
	logFile, err := os.OpenFile(logPath, logFileFlags, logFilePerm)
	if err != nil {
		return nil, err
	}

	// 4. Determine stdout logging behavior.
	logStdOut := true
	if val, ok := os.LookupEnv("ENC_SERVER_GO_LOG_STDOUT"); ok {
		if parsed, err := strconv.ParseBool(val); err == nil {
			logStdOut = parsed
		}
	}

	// 5. Multi-writer target (file + optional stdout).
	var writer io.Writer = logFile
	if logStdOut {
		writer = io.MultiWriter(os.Stdout, logFile)
	}

	// 6. Build structured JSON handler with service metadata.
	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format time to guarantee exact 6-digit microsecond width (padded with zeros)
			if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
				formattedTime := a.Value.Time().Format("2006-01-02T15:04:05.000000Z07:00")
				return slog.String(slog.TimeKey, formattedTime)
			}
			return a
		},
	}
	jsonHandler := slog.NewJSONHandler(writer, handlerOpts)

	// Create logger with pre-attached service attribute
	logger := slog.New(jsonHandler).With(slog.String("service", serviceName))

	// Set as global default logger
	slog.SetDefault(logger)

	return logFile, nil
}
