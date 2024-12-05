package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"os"
	"path/filepath"
	"time"
)

var (
	// defaultLogger is the default logger instance
	defaultLogger zerolog.Logger
	consoleWriter zerolog.ConsoleWriter
	fileWriters   []io.Writer

	// ANSI color codes
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorReset   = "\033[0m"
)

func init() {
	// Set default time format to RFC3339 with millisecond precision
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Create console writer with custom formatting
	consoleWriter = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05.000",
		FormatLevel: func(i interface{}) string {
			level := fmt.Sprintf("%s", i)
			switch level {
			case "debug":
				return fmt.Sprintf("%s[DEBUG]%s", colorBlue, colorReset)
			case "info":
				return fmt.Sprintf("%s[INFO]%s", colorGreen, colorReset)
			case "warn":
				return fmt.Sprintf("%s[WARN]%s", colorYellow, colorReset)
			case "error":
				return fmt.Sprintf("%s[ERROR]%s", colorRed, colorReset)
			case "fatal":
				return fmt.Sprintf("%s[FATAL]%s", colorMagenta, colorReset)
			default:
				return fmt.Sprintf("%s[%s]%s", colorCyan, level, colorReset)
			}
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("| %s", i)
		},
		FormatFieldName: func(i interface{}) string {
			return fmt.Sprintf("%s%s%s:", colorCyan, i, colorReset)
		},
		FormatFieldValue: func(i interface{}) string {
			return fmt.Sprintf("%s%v%s", colorBlue, i, colorReset)
		},
		FormatTimestamp: func(i interface{}) string {
			return fmt.Sprintf("%s%s%s |", colorGreen, i, colorReset)
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

	// Add file to writers list
	fileWriters = append(fileWriters, file)

	// Create multi-writer for both console and file outputs
	writers := make([]io.Writer, 0, len(fileWriters)+1)
	writers = append(writers, consoleWriter)
	for _, fw := range fileWriters {
		writers = append(writers, fw)
	}
	multiWriter := zerolog.MultiLevelWriter(writers...)

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
