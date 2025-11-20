package logging

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger holds the global logger instance
var Logger zerolog.Logger

// Initialize sets up the global logger with the specified configuration
func Initialize(debug bool, output io.Writer) {
	// Set the global log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
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

// GetLogger returns a logger with a specific component name
func GetLogger(component string) zerolog.Logger {
	return Logger.With().Str("component", component).Logger()
}

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled() bool {
	return zerolog.GlobalLevel() <= zerolog.DebugLevel
}
