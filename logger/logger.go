package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
	"time"
)

var (
	// defaultLogger is the default logger instance
	defaultLogger zerolog.Logger
)

func init() {
	// Set default time format to RFC3339 with millisecond precision
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Create console writer
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05.000",
		FormatLevel: func(i interface{}) string {
			return fmt.Sprintf("| %-6s|", i)
		},
	}

	// Initialize the default logger
	defaultLogger = zerolog.New(consoleWriter).
		With().
		Timestamp().
		Logger()
}

// SetLogLevel sets the global log level
func SetLogLevel(level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
}

// AddFileOutput adds a log file output
func AddFileOutput(logPath string) error {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// Create multi-writer for both console and file
	multiWriter := zerolog.MultiLevelWriter(file)

	// Update default logger with new writer
	defaultLogger = defaultLogger.Output(multiWriter)
	return nil
}

// Fields type for structured logging
type Fields map[string]interface{}

// log creates a new event with fields
func log(level zerolog.Level, msg string, err error, fields Fields) {
	event := defaultLogger.WithLevel(level)

	if err != nil {
		event = event.Err(err)
	}

	for k, v := range fields {
		event = event.Interface(k, v)
	}

	event.Msg(msg)
}

// Debug logs a debug message
func Debug(msg string, fields Fields) {
	log(zerolog.DebugLevel, msg, nil, fields)
}

// Info logs an info message
func Info(msg string, fields Fields) {
	log(zerolog.InfoLevel, msg, nil, fields)
}

// Warn logs a warning message
func Warn(msg string, fields Fields) {
	log(zerolog.WarnLevel, msg, nil, fields)
}

// Error logs an error message
func Error(msg string, err error, fields Fields) {
	log(zerolog.ErrorLevel, msg, err, fields)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, err error, fields Fields) {
	log(zerolog.FatalLevel, msg, err, fields)
	os.Exit(1)
}

// GetLogger returns the default logger instance
func GetLogger() *zerolog.Logger {
	return &defaultLogger
}

// WithFields creates a new logger with the given fields
func WithFields(fields Fields) *zerolog.Logger {
	logger := defaultLogger.With().Fields(fields).Logger()
	return &logger
}
