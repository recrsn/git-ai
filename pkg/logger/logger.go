package logger

import (
	golog "log"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// DEBUG level for verbose debug information
	DEBUG LogLevel = iota
	// INFO level for general information
	INFO
	// WARN level for warning messages
	WARN
	// ERROR level for error messages
	ERROR
	// FATAL level for critical errors that should terminate the program
	FATAL
)

var (
	// currentLevel controls logging verbosity
	currentLevel = FATAL
)

// levelPrefixes maps log levels to text prefixes
var levelPrefixes = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO ",
	WARN:  "WARN ",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// SetLevel sets the current log level
func SetLevel(verbosity int) {
	if verbosity < 0 {
		verbosity = 0
	}
	if verbosity > int(FATAL) {
		verbosity = int(FATAL)
	}
	currentLevel = FATAL - LogLevel(verbosity)
}

// log prints a log message with the specified level
func log(level LogLevel, format string, args ...interface{}) {
	if level < currentLevel {
		return
	}
	prefix := levelPrefixes[level]
	formatWithLogInfo := prefix + ": " + format + "\n"

	if level == FATAL {
		golog.Fatalf(formatWithLogInfo, args...)
	} else {
		golog.Printf(formatWithLogInfo, args...)
	}
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	log(DEBUG, format, args...)
}

// Info logs an informational message
func Info(format string, args ...interface{}) {
	log(INFO, format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	log(WARN, format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	log(ERROR, format, args...)
}

// Fatal logs a fatal error message and exits the program
func Fatal(format string, args ...interface{}) {
	log(FATAL, format, args...)
}
