// Package logger provides a production-grade structured logger (zap) with
// OpenTelemetry trace correlation, so logs, metrics and traces line up in
// Grafana.
package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a structured logger tagged with service metadata.
// format is "json" (production) or "console" (local development).
func New(level, format, service, version, environment string) (*zap.Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		lvl = zapcore.InfoLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.MessageKey = "message"

	if format == "console" {
		cfg.Encoding = "console"
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return l.With(
		zap.String("service", service),
		zap.String("version", version),
		zap.String("env", environment),
	), nil
}

// WithTrace enriches a logger with trace and span IDs from ctx, enabling
// Grafana to navigate from a log line straight to its trace.
func WithTrace(ctx context.Context, l *zap.Logger) *zap.Logger {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return l
	}
	return l.With(
		zap.String("trace_id", sc.TraceID().String()),
		zap.String("span_id", sc.SpanID().String()),
	)
}
