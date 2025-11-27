package logging

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger holds the global logger instance.
var Logger zerolog.Logger

func init() {
	// Initialize with a default logger to stderr
	// This ensures logging works even before Initialize() is called
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}
	Logger = zerolog.New(consoleWriter).
		With().
		Timestamp().
		Caller().
		Logger()
	log.Logger = Logger
	// Default: only show errors and fatal messages
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}

// Initialize sets up the global logger with the specified configuration.
func Initialize(debug bool, output io.Writer) {
	// Set the global log level
	// Default: only show errors (quiet mode)
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	if debug {
		// Debug mode: show everything
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Use console writer for human-readable output in development
	if output == nil {
		output = os.Stderr
	}

	// Configure console writer for pretty output
	consoleWriter := zerolog.ConsoleWriter{
		Out:        output,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	// Create the global logger
	Logger = zerolog.New(consoleWriter).
		With().
		Timestamp().
		Caller().
		Logger()

	// Update the global logger
	log.Logger = Logger

	if debug {
		Logger.Debug().Msg("Debug mode enabled")
	}
}

// GetLogger returns a logger with a specific component name.
func GetLogger(component string) *zerolog.Logger {
	logger := Logger.With().Str("component", component).Logger()
	return &logger
}

// IsDebugEnabled returns true if debug logging is enabled.
func IsDebugEnabled() bool {
	return zerolog.GlobalLevel() <= zerolog.DebugLevel
}

// RedirectLoggingToFile redirects logging to a file during interactive UI modes.
// This prevents logs from interfering with the TUI while still capturing them.
// Returns the log file and the previous logger so they can be restored later.
func RedirectLoggingToFile() (*os.File, zerolog.Logger, error) {
	// Create a temporary log file
	tmpDir := os.TempDir()
	logFile, err := os.CreateTemp(tmpDir, "monhang-*.log")
	if err != nil {
		return nil, Logger, err
	}

	// Save the previous logger
	previousLogger := Logger

	// Create a new logger that writes to the file
	fileLogger := zerolog.New(logFile).
		With().
		Timestamp().
		Caller().
		Logger()

	// Update the global logger
	Logger = fileLogger
	log.Logger = fileLogger

	return logFile, previousLogger, nil
}

// RestoreLogger restores the logger to a previous instance and closes the log file.
func RestoreLogger(logFile *os.File, previousLogger zerolog.Logger) string {
	logPath := ""
	if logFile != nil {
		logPath = logFile.Name()
		_ = logFile.Close() // Ignore error on close
	}

	// Restore the previous logger
	Logger = previousLogger
	log.Logger = previousLogger

	return logPath
}
