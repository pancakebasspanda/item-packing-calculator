package main

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace/noop"
)

// setupTelemetry initialises the OpenTelemetry provider and exporter.
func setupTelemetry(ctx context.Context) *trace.TracerProvider {

	// If otelEndpoint is empty or the exporter cannot be created, it falls back to
	// a no-op provider so the server starts cleanly without Jaeger available.
	if otelEndpoint == "" {
		logger.Warn().Msg("otel-endpoint not set — tracing disabled, using no-op provider")
		otel.SetTracerProvider(noop.NewTracerProvider())
		return nil
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(otelEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to create otel exporter — tracing disabled, using no-op provider")
		otel.SetTracerProvider(noop.NewTracerProvider())
		return nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(_appName),
		),
	)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to create OTel resource — tracing disabled, using no-op provider")
		otel.SetTracerProvider(noop.NewTracerProvider())
		return nil
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	// set the global provider so handlers can access it later to create spans
	otel.SetTracerProvider(tracerProvider)

	logger.Info().Str("endpoint", otelEndpoint).Msg("otel tracing enabled")

	return tracerProvider
}
