package logger

import (
	"fmt"
	"log"
	"os"
)

// Logger wraps standard log with enable/disable capability
type Logger struct {
	enabled bool
}

// global logger instance
var globalLogger = &Logger{enabled: true}

// SetEnabled enables or disables all logging output
func SetEnabled(enabled bool) {
	globalLogger.enabled = enabled
}

// Enabled returns whether logging is enabled
func Enabled() bool {
	return globalLogger.enabled
}

// Print prints a message if logging is enabled
func Print(v ...interface{}) {
	if globalLogger.enabled {
		log.Print(v...)
	}
}

// Println prints a message with newline if logging is enabled
func Println(v ...interface{}) {
	if globalLogger.enabled {
		log.Println(v...)
	}
}

// Printf prints a formatted message if logging is enabled
func Printf(format string, v ...interface{}) {
	if globalLogger.enabled {
		log.Printf(format, v...)
	}
}

// Fatal prints a message and exits (always executes, cannot be disabled)
func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

// Fatalf prints a formatted message and exits (always executes, cannot be disabled)
func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

// Fatalln prints a message with newline and exits (always executes, cannot be disabled)
func Fatalln(v ...interface{}) {
	log.Fatalln(v...)
}

// PrintError prints an error message (used in middleware, respects enabled flag)
func PrintError(format string, v ...interface{}) {
	if globalLogger.enabled {
		fmt.Fprintf(os.Stderr, format, v...)
	}
}

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var level = INFO

func SetLevel(l Level) { level = l }

func levelToString(l Level) string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (l *Logger) logf(lvl Level, format string, v ...interface{}) {
	if !l.enabled || lvl < level {
		return
	}
	log.Output(3, fmt.Sprintf("[%s] %s", levelToString(lvl), fmt.Sprintf(format, v...)))
}

func Debugf(format string, v ...interface{})   { globalLogger.logf(DEBUG, format, v...) }
func Infof(format string, v ...interface{})    { globalLogger.logf(INFO, format, v...) }
func Warnf(format string, v ...interface{})    { globalLogger.logf(WARN, format, v...) }
func Errorf(format string, v ...interface{})   { globalLogger.logf(ERROR, format, v...) }
