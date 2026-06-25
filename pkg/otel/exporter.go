package otel

import (
	"context"
	"crypto/tls"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const (
	serviceName         = "lite-api"
	defaultOTLPEndpoint = "localhost:4318"
)

func Enabled() bool {
	enabled := strings.TrimSpace(strings.ToLower(os.Getenv("OTEL_ENABLED")))
	return enabled == "true" || enabled == "1" || enabled == "yes"
}

func Endpoint() string {
	endpoint := strings.TrimSpace(os.Getenv("OTEL_ENDPOINT"))
	if endpoint != "" {
		return endpoint
	}
	return defaultOTLPEndpoint
}

func SetupTracer() (func(context.Context) error, error) {
	if !Enabled() {
		return func(context.Context) error { return nil }, nil
	}

	ctx := context.Background()
	return InstallExportPipeline(ctx)
}

func Resource() *resource.Resource {
	// Defines resource with service name, version, and environment.
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(os.Getenv("VERSION")),
		attribute.String("environment", os.Getenv("ENV")),
	)
}

func InstallExportPipeline(ctx context.Context) (func(context.Context) error, error) {
	var tlsOption otlptracehttp.Option
	if os.Getenv("ENV") == "dev" {
		tlsOption = otlptracehttp.WithInsecure()
	} else {
		tlsOption = otlptracehttp.WithTLSClientConfig(&tls.Config{})
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(Endpoint()),
		tlsOption,
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(Resource()),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tracerProvider.Shutdown, nil
}
