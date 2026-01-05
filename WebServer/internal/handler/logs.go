package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"webserver/internal/config"
	"webserver/internal/logger"
)

// ShowInfoLogsHandler serves the info.log file as text/plain.
func ShowInfoLogsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveLogFile(w, r, cfg.LogDirectory, "info.log")
	}
}

// ShowWarningLogsHandler serves the warning.log file as text/plain.
func ShowWarningLogsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveLogFile(w, r, cfg.LogDirectory, "warning.log")
	}
}

// ShowErrorLogsHandler serves the error.log file as text/plain.
func ShowErrorLogsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveLogFile(w, r, cfg.LogDirectory, "error.log")
	}
}

// serveLogFile is a helper that sets headers and serves a log file if it exists.
func serveLogFile(w http.ResponseWriter, r *http.Request, logDir, filename string) {
	filePath := filepath.Join(logDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Log file not found: " + filename))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")

	http.ServeFile(w, r, filePath)
}

// ClearInfoLogsHandler truncates info.log via the logger utility.
func ClearInfoLogsHandler(logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.CleanLogs("info.log")
	}
}

// ClearWarningLogsHandler truncates warning.log via the logger utility.
func ClearWarningLogsHandler(logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.CleanLogs("warning.log")
	}
}

// ClearErrorLogsHandler truncates error.log via the logger utility.
func ClearErrorLogsHandler(logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.CleanLogs("error.log")
	}
}
