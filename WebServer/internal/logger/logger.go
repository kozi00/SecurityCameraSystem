package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"webserver/internal/config"
)

// Logger struktura główna loggera
type Logger struct {
	infoLog    *log.Logger
	warningLog *log.Logger
	errorLog   *log.Logger
	logDir     string
	mu         sync.Mutex
}

func NewLogger(config *config.Config) *Logger {
	// Utwórz katalog na logi jeśli nie istnieje
	if err := os.MkdirAll(config.LogDirectory, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	logger := &Logger{
		logDir: config.LogDirectory,
	}

	logger.setupLoggers()
	return logger
}

// setupLoggers konfiguruje poszczególne loggery
func (l *Logger) setupLoggers() {

	// Ścieżki do plików logów
	infoFile := filepath.Join(l.logDir, "info.log")
	warningFile := filepath.Join(l.logDir, "warning.log")
	errorFile := filepath.Join(l.logDir, "error.log")

	// Otwórz pliki logów
	infoFileHandle := l.openLogFile(infoFile)
	warningFileHandle := l.openLogFile(warningFile)
	errorFileHandle := l.openLogFile(errorFile)

	// Utwórz multi-writery (konsola + plik)
	infoWriter := io.MultiWriter(os.Stdout, infoFileHandle)
	warningWriter := io.MultiWriter(os.Stdout, warningFileHandle)
	errorWriter := io.MultiWriter(os.Stderr, errorFileHandle)

	// Skonfiguruj loggery z prefiksami
	l.infoLog = log.New(infoWriter, "ℹ️  INFO    ", log.Ldate|log.Ltime|log.Lshortfile)
	l.warningLog = log.New(warningWriter, "⚠️  WARNING ", log.Ldate|log.Ltime|log.Lshortfile)
	l.errorLog = log.New(errorWriter, "❌ ERROR   ", log.Ldate|log.Ltime|log.Lshortfile)
}

// openLogFile otwiera plik loga w trybie append
func (l *Logger) openLogFile(filename string) *os.File {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file %s: %v", filename, err)
	}
	return file
}

// Info loguje wiadomość informacyjną
func (l *Logger) Info(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.infoLog.Printf(format, v...)
}

// Warning loguje ostrzeżenie
func (l *Logger) Warning(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.warningLog.Printf(format, v...)
}

// Error loguje błąd
func (l *Logger) Error(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errorLog.Printf(format, v...)
}

// CleanOldLogs usuwa stare pliki logów (starsze niż 'days' dni)
func (l *Logger) CleanLogs(fileName string) {
	filePath := filepath.Join(l.logDir, fileName)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		l.Error("Błąd przy otwieraniu pliku: %v", err)
	}
	defer file.Close()

	l.Info("Zawartość pliku została usunięta.")
}

// GetLogStats zwraca statystyki logów
// func (l *Logger) GetLogStats() map[string]interface{} {
// 	files, _ := filepath.Glob(filepath.Join(l.logDir, "*.log"))

// 	stats := make(map[string]interface{})
// 	stats["log_directory"] = l.logDir
// 	stats["total_log_files"] = len(files)

// 	var totalSize int64
// 	for _, file := range files {
// 		if info, err := os.Stat(file); err == nil {
// 			totalSize += info.Size()
// 		}
// 	}

// 	stats["total_size_mb"] = float64(totalSize) / (1024 * 1024)
// 	return stats
// }
