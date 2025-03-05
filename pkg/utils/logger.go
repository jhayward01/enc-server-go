package utils

import (
	"io"
	"io/fs"
	"log"
	"os"
	"strconv"
	"time"
)

const defaultDir = "/tmp/enc-server-go-logs"

func StartLog(name string) (logFile *os.File, err error) {

	// Load log directory override.
	logDir := defaultDir
	if val, ok := os.LookupEnv("ENC_SERVER_GO_LOG_DIR"); ok {
		logDir = val
	}

	// Create directory folder
	if err = os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	// Create log file.
	timestamp := time.Now().UTC().Format(time.RFC3339)
	logPath := logDir + "/" + name + "." + timestamp + ".log"
	logFileFlags := os.O_CREATE | os.O_APPEND | os.O_RDWR
	logFilePerm := fs.FileMode(0666)
	if logFile, err = os.OpenFile(logPath, logFileFlags, logFilePerm); err != nil {
		return nil, err
	}

	// Load standard out logging override.
	logStdOut := true
	if val, ok := os.LookupEnv("ENC_SERVER_GO_LOG_STDOUT"); ok {
		if logStdOut, err = strconv.ParseBool(val); err != nil {
			return nil, err
		}
	}

	// Redirect logging to log file and standard output.
	o := io.Writer(logFile)
	if logStdOut {
		o = io.MultiWriter(os.Stdout, logFile)
	}
	log.SetOutput(o)

	// Misc. logger configuration
	logFlags := log.Ldate | log.Ltime | log.LUTC
	log.SetFlags(logFlags)

	return logFile, nil
}
