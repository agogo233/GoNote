package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type appLogger struct {
	enabled bool
	json    bool
	logger  *log.Logger
}

var globalLogger = &appLogger{
	logger: log.New(os.Stderr, "", 0),
}

func SetEnabled(e bool) { globalLogger.enabled = e }

func SetJSONOutput(j bool) { globalLogger.json = j }

func Enabled() bool {
	return globalLogger.enabled
}

func Print(v ...interface{}) {
	if globalLogger.enabled {
		if globalLogger.json {
			writeJSON("INFO", fmt.Sprint(v...))
		} else {
			globalLogger.logger.Print(v...)
		}
	}
}

func Println(v ...interface{}) {
	if globalLogger.enabled {
		if globalLogger.json {
			writeJSON("INFO", fmt.Sprint(v...))
		} else {
			globalLogger.logger.Println(v...)
		}
	}
}

func Printf(format string, v ...interface{}) {
	if globalLogger.enabled {
		if globalLogger.json {
			writeJSON("INFO", fmt.Sprintf(format, v...))
		} else {
			globalLogger.logger.Printf(format, v...)
		}
	}
}

func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	if globalLogger.json {
		writeJSON("FATAL", fmt.Sprintf(format, v...))
	}
	log.Fatalf(format, v...)
}

func Fatalln(v ...interface{}) {
	log.Fatalln(v...)
}

func PrintError(format string, v ...interface{}) {
	if globalLogger.enabled {
		if globalLogger.json {
			writeJSON("ERROR", fmt.Sprintf(format, v...))
		} else {
			globalLogger.logger.Printf("[ERROR] "+format, v...)
		}
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

func (l *appLogger) logf(lvl Level, format string, v ...interface{}) {
	if !l.enabled || lvl < level {
		return
	}
	if l.json {
		writeJSON(levelToString(lvl), fmt.Sprintf(format, v...))
	} else {
		log.Output(3, fmt.Sprintf("[%s] %s", levelToString(lvl), fmt.Sprintf(format, v...)))
	}
}

func Debugf(format string, v ...interface{})   { globalLogger.logf(DEBUG, format, v...) }
func Infof(format string, v ...interface{})    { globalLogger.logf(INFO, format, v...) }
func Warnf(format string, v ...interface{})    { globalLogger.logf(WARN, format, v...) }
func Errorf(format string, v ...interface{})   { globalLogger.logf(ERROR, format, v...) }

func writeJSON(level, msg string) {
	entry := struct {
		Level   string `json:"level"`
		Message string `json:"message"`
		Time    string `json:"time"`
	}{
		Level:   level,
		Message: msg,
		Time:    time.Now().Format(time.RFC3339),
	}
	data, _ := json.Marshal(entry)
	globalLogger.logger.Println(string(data))
}
