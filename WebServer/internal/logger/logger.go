package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"webserver/internal/config"
)

// Logger provides leveled logging (info/warning/error) to files and stdout/stderr.
type Logger struct {
	infoLog    *log.Logger
	warningLog *log.Logger
	errorLog   *log.Logger
	logDir     string
	mu         sync.Mutex
}

// NewLogger creates a Logger and ensures the log directory exists.
func NewLogger(config *config.Config) *Logger {
	if err := os.MkdirAll(config.LogDirectory, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	logger := &Logger{
		logDir: config.LogDirectory,
	}

	logger.setupLoggers()
	return logger
}

// setupLoggers initializes writers and per-level loggers.
func (l *Logger) setupLoggers() {
	infoFile := filepath.Join(l.logDir, "info.log")
	warningFile := filepath.Join(l.logDir, "warning.log")
	errorFile := filepath.Join(l.logDir, "error.log")

	infoFileHandle := l.openLogFile(infoFile)
	warningFileHandle := l.openLogFile(warningFile)
	errorFileHandle := l.openLogFile(errorFile)

	infoWriter := io.MultiWriter(os.Stdout, infoFileHandle)
	warningWriter := io.MultiWriter(os.Stdout, warningFileHandle)
	errorWriter := io.MultiWriter(os.Stderr, errorFileHandle)

	l.infoLog = log.New(infoWriter, "ℹ️  INFO    ", log.Ldate|log.Ltime|log.Lshortfile)
	l.warningLog = log.New(warningWriter, "⚠️  WARNING ", log.Ldate|log.Ltime|log.Lshortfile)
	l.errorLog = log.New(errorWriter, "❌ ERROR   ", log.Ldate|log.Ltime|log.Lshortfile)
}

// openLogFile opens or creates a log file for appending.
func (l *Logger) openLogFile(filename string) *os.File {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file %s: %v", filename, err)
	}
	return file
}

// Info writes a formatted info-level log entry.
func (l *Logger) Info(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.infoLog.Printf(format, v...)
}

// Warning writes a formatted warning-level log entry.
func (l *Logger) Warning(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.warningLog.Printf(format, v...)
}

// Error writes a formatted error-level log entry.
func (l *Logger) Error(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errorLog.Printf(format, v...)
}

// CleanLogs truncates the specified log file.
func (l *Logger) CleanLogs(fileName string) {
	filePath := filepath.Join(l.logDir, fileName)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		l.Error("Error opening file: %v", err)
	}
	defer file.Close()

	l.Info("File content has been cleared.")
}
