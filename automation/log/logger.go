package log

import (
	"context"
	"log/slog"
	"os"
)

// Level mirrors common severity including Trace below Debug.
type Level int

const (
	LevelTrace Level = -8
	LevelDebug Level = Level(slog.LevelDebug)
	LevelInfo  Level = Level(slog.LevelInfo)
	LevelWarn  Level = Level(slog.LevelWarn)
	LevelError Level = Level(slog.LevelError)
)

// Logger is the framework logging surface. Prefer this over fmt printing.
type Logger interface {
	Trace(msg string, args ...any)
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

type slogLogger struct {
	s *slog.Logger
}

// NewSlog wraps a slog.Logger. If s is nil, a text handler to stderr at Info is used.
func NewSlog(s *slog.Logger) Logger {
	if s == nil {
		s = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return &slogLogger{s: s}
}

// Nop returns a logger that discards all output.
func Nop() Logger { return nop{} }

type nop struct{}

func (nop) Trace(string, ...any) {}
func (nop) Debug(string, ...any) {}
func (nop) Info(string, ...any)  {}
func (nop) Warn(string, ...any)  {}
func (nop) Error(string, ...any) {}
func (nop) With(...any) Logger   { return nop{} }

func (l *slogLogger) Trace(msg string, args ...any) {
	l.s.Log(context.Background(), slog.Level(LevelTrace), msg, args...)
}
func (l *slogLogger) Debug(msg string, args ...any) { l.s.Debug(msg, args...) }
func (l *slogLogger) Info(msg string, args ...any)  { l.s.Info(msg, args...) }
func (l *slogLogger) Warn(msg string, args ...any)  { l.s.Warn(msg, args...) }
func (l *slogLogger) Error(msg string, args ...any) { l.s.Error(msg, args...) }
func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{s: l.s.With(args...)}
}
