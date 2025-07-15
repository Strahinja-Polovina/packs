package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	level  Level
	logger *log.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = New(INFO)
}

func New(level Level) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] [%s] ", timestamp, level.String())
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("%s%s", prefix, message)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

// Global functions using default logger
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

func SetLevel(level Level) {
	defaultLogger.level = level
}

func GetLogger() *Logger {
	return defaultLogger
}
