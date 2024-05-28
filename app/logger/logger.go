package logger

import (
	"fmt"
	"log"
)

type LogLevel int

const (
	LOG_LEVEL_DEBUG LogLevel = iota
	LOG_LEVEL_INFO
	LOG_LEVEL_WARN
	LOG_LEVEL_ERROR
)

var levelMap = map[LogLevel]string{
	LOG_LEVEL_DEBUG: "DEBUG",
	LOG_LEVEL_INFO:  "INFO ",
	LOG_LEVEL_WARN:  "WARN ",
	LOG_LEVEL_ERROR: "ERROR",
}

type CustomLogger struct {
	level       LogLevel
	handlerFunc func(string, LogLevel) string
}

func NewLogger(port int, loggerLevel LogLevel) *CustomLogger {
	format := "[%s] [localhost:%04d]"
	handler := func(s string, level LogLevel) string {
		return fmt.Sprintf(format, levelMap[level], port) + " " + s
	}
	return &CustomLogger{loggerLevel, handler}
}

func (logger *CustomLogger) Logf(level LogLevel, s string, v ...any) {
	if level < logger.level {
		return
	}
	if len(v) == 0 {
		log.Print(logger.handlerFunc(s, level))
		return
	}
	if len(s) > 0 && s[len(s)-1] != '\n' {
		s += "\n"
	}
	log.Printf(logger.handlerFunc(s, level), v...)
}

func (logger *CustomLogger) Log(level LogLevel, s string) {
	logger.Logf(level, s)
}

func (logger *CustomLogger) Debug(s string) {
	logger.Log(LOG_LEVEL_DEBUG, s)
}

func (logger *CustomLogger) Info(s string) {
	logger.Log(LOG_LEVEL_INFO, s)
}

func (logger *CustomLogger) Warn(s string) {
	logger.Log(LOG_LEVEL_WARN, s)
}

func (logger *CustomLogger) Error(s string) {
	logger.Log(LOG_LEVEL_ERROR, s)
}

func (logger *CustomLogger) Debugf(s string, v ...any) {
	logger.Logf(LOG_LEVEL_DEBUG, s, v...)
}

func (logger *CustomLogger) Infof(s string, v ...any) {
	logger.Logf(LOG_LEVEL_INFO, s, v...)
}

func (logger *CustomLogger) Warnf(s string, v ...any) {
	logger.Logf(LOG_LEVEL_WARN, s, v...)
}

func (logger *CustomLogger) Errorf(s string, v ...any) {
	logger.Logf(LOG_LEVEL_ERROR, s, v...)
}
