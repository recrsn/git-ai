package logger

import (
	"fmt"
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
	// CurrentLevel is the current log level
	CurrentLevel = INFO
)

// levelColors maps log levels to ANSI color codes
var levelColors = map[LogLevel]string{
	DEBUG: "\033[36m", // Cyan
	INFO:  "\033[32m", // Green
	WARN:  "\033[33m", // Yellow
	ERROR: "\033[31m", // Red
	FATAL: "\033[35m", // Magenta
}

// levelPrefixes maps log levels to text prefixes
var levelPrefixes = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO ",
	WARN:  "WARN ",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// resetColor is the ANSI escape code to reset text color
const resetColor = "\033[0m"

// SetLevel sets the current log level
func SetLevel(level LogLevel) {
	CurrentLevel = level
}

// log prints a log message with the specified level
func log(level LogLevel, format string, args ...interface{}) {
	if level < CurrentLevel {
		return
	}

	prefix := levelPrefixes[level]
	color := levelColors[level]

	message := fmt.Sprintf(format, args...)
	if level == FATAL {
		golog.Fatalf("%s%s: %s%s\n", color, prefix, message, resetColor)
	} else {
		golog.Printf("%s%s: %s%s\n", color, prefix, message, resetColor)
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

// PrintMessage prints an unformatted message to stdout (for user-facing output)
func PrintMessage(message string) {
	fmt.Println(message)
}

// PrintMessagef prints a formatted message to stdout (for user-facing output)
func PrintMessagef(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
