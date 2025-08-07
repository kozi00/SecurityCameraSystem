package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"webserver/internal/config"
	"webserver/internal/logger"
)

func ShowInfoLogsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveLogFile(w, r, cfg.LogDirectory, "info.log")
	}
}

// ✅ SERVE WARNING LOGS
func ShowWarningLogsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveLogFile(w, r, cfg.LogDirectory, "warning.log")
	}
}

// ✅ SERVE ERROR LOGS
func ShowErrorLogsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveLogFile(w, r, cfg.LogDirectory, "error.log")
	}
}

// ✅ HELPER: Serve single log file
func serveLogFile(w http.ResponseWriter, r *http.Request, logDir, filename string) {
	filePath := filepath.Join(logDir, filename)

	// Sprawdź czy plik istnieje
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Log file not found: " + filename))
		return
	}

	// Ustaw nagłówki
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")

	// Serwuj plik
	http.ServeFile(w, r, filePath)
}

func ClearInfoLogsHandler(logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.CleanLogs("info.log")
	}
}
func ClearWarningLogsHandler(logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.CleanLogs("warning.log")
	}
}
func ClearErrorLogsHandler(logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.CleanLogs("error.log")
	}
}
