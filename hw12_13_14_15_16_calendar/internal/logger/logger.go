package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type level int

const (
	DEBUG level = iota
	INFO
	WARN
	ERROR
)

func parseLevel(s string) level {
	switch strings.ToLower(s) {
	case "debug":
		return DEBUG
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

type Logger struct {
	out   io.Writer
	level level
	mu    sync.Mutex
}

func New(levelStr string) *Logger {
	return &Logger{
		out:   os.Stdout,
		level: parseLevel(levelStr),
	}
}

func (l *Logger) log(lv level, prefix, msg string) {
	if lv < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	ts := time.Now().Format(time.RFC3339)
	fmt.Fprintf(l.out, "%s [%s] %s\n", ts, prefix, msg)
}

func (l *Logger) Info(msg string) {
	l.log(INFO, "INFO", msg)
}

func (l *Logger) Error(msg string) {
	l.log(ERROR, "ERROR", msg)
}

func (l *Logger) Debug(msg string) {
	l.log(DEBUG, "DEBUG", msg)
}

func (l *Logger) Warn(msg string) {
	l.log(WARN, "WARN", msg)
}
