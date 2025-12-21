package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Logger handles simple file-based logging
type Logger struct {
	file *os.File
}

// NewLogger creates/opens a log file in config dir
func NewLogger() (*Logger, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	appDir := filepath.Join(configDir, "pdf-freezer")
	logPath := filepath.Join(appDir, "app.log")

	// Append mode
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{file: f}, nil
}

// Info logs connection info
func (l *Logger) Info(msg string) {
	l.write("INFO", msg)
}

// Error logs error
func (l *Logger) Error(msg string) {
	l.write("ERROR", msg)
}

func (l *Logger) write(level, msg string) {
	ts := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("[%s] %s: %s\n", ts, level, msg)
	if l.file != nil {
		l.file.WriteString(line)
	}
	// Also print to stdout for Wails dev console
	fmt.Print(line)
}

// Close closes the file
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}
