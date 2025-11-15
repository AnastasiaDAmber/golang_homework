package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerLevels(t *testing.T) {
	buf := &bytes.Buffer{}
	l := &Logger{out: buf, level: parseLevel("info")}

	l.Debug("should not appear")
	l.Info("hello")
	l.Error("oops")

	s := buf.String()
	if strings.Contains(s, "should not appear") {
		t.Fatalf("debug message printed when level=info")
	}
	if !strings.Contains(s, "hello") || !strings.Contains(s, "oops") {
		t.Fatalf("expected messages missing: %s", s)
	}
}
