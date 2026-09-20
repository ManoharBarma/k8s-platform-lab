// Package telemetry configures OpenTelemetry for the platform: a global tracer
// provider and meter provider that export via OTLP/gRPC to the OpenTelemetry
// Collector. Trace-context and baggage propagators are always installed so
// trace IDs flow across REST, gRPC and Kafka boundaries.
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config controls telemetry setup.
type Config struct {
	Enabled          bool
	ServiceName      string
	ServiceVersion   string
	Environment      string
	OTLPEndpoint     string
	Insecure         bool
	TraceSampleRatio float64
}

// Provider bundles the configured OTel providers so they can be shut down
// cleanly on service exit (flushing any buffered spans/metrics).
type Provider struct {
	shutdownFns []func(context.Context) error
}

// Shutdown flushes and stops all telemetry pipelines.
func (p *Provider) Shutdown(ctx context.Context) error {
	var err error
	for i := len(p.shutdownFns) - 1; i >= 0; i-- {
		err = errors.Join(err, p.shutdownFns[i](ctx))
	}
	return err
}

// Setup installs global propagators and, when enabled, OTLP trace and metric
// providers. It is always safe to call: when disabled it installs only the
// propagators and returns a no-op Provider.
func Setup(ctx context.Context, cfg Config) (*Provider, error) {
	p := &Provider{}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if !cfg.Enabled {
		return p, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("deployment.environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}

	// --- Traces -> OTLP/gRPC (Collector -> Tempo) ---
	traceOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint)}
	if cfg.Insecure {
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
	}
	traceExp, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		return nil, fmt.Errorf("otlp trace exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.TraceSampleRatio))),
	)
	otel.SetTracerProvider(tp)
	p.shutdownFns = append(p.shutdownFns, tp.Shutdown)

	// --- Metrics -> OTLP/gRPC (Collector -> Prometheus) ---
	metricOpts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(cfg.OTLPEndpoint)}
	if cfg.Insecure {
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
	}
	metricExp, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		return nil, fmt.Errorf("otlp metric exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp,
			sdkmetric.WithInterval(15*time.Second))),
	)
	otel.SetMeterProvider(mp)
	p.shutdownFns = append(p.shutdownFns, mp.Shutdown)

	return p, nil
}
