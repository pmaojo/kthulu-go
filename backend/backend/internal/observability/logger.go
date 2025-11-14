package observability

import (
	"backend/core"
	"go.uber.org/zap"
)

// Logger defines the interface used by middlewares and services.
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Sync() error
}

// ZapLogger implements Logger using a zap.Logger.
type ZapLogger struct {
	*zap.Logger
}

// NewLogger constructs a zap based logger from configuration.
func NewLogger(cfg *core.Config) (Logger, error) {
	z, err := core.NewZapLogger(cfg)
	if err != nil {
		return nil, err
	}
	return &ZapLogger{z}, nil
}

// NewNopLogger returns a logger that does nothing.
func NewNopLogger() Logger {
	return &ZapLogger{zap.NewNop()}
}

func (l *ZapLogger) Debug(msg string, fields ...zap.Field) { l.Logger.Debug(msg, fields...) }
func (l *ZapLogger) Info(msg string, fields ...zap.Field)  { l.Logger.Info(msg, fields...) }
func (l *ZapLogger) Warn(msg string, fields ...zap.Field)  { l.Logger.Warn(msg, fields...) }
func (l *ZapLogger) Error(msg string, fields ...zap.Field) { l.Logger.Error(msg, fields...) }
func (l *ZapLogger) Fatal(msg string, fields ...zap.Field) { l.Logger.Fatal(msg, fields...) }
func (l *ZapLogger) With(fields ...zap.Field) Logger       { return &ZapLogger{l.Logger.With(fields...)} }
func (l *ZapLogger) Sync() error                           { return l.Logger.Sync() }

// GetZapLogger extracts the underlying zap.Logger when required.
func GetZapLogger(l Logger) *zap.Logger {
	if zl, ok := l.(*ZapLogger); ok {
		return zl.Logger
	}
	return zap.NewNop()
}
