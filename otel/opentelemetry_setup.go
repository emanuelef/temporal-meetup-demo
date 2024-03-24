package opentelemetry

import (
	"context"
	"log"

	_ "github.com/joho/godotenv/autoload"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func InitializeGlobalTracerProvider(ctx context.Context) (*sdktrace.TracerProvider, error) {
	// Initialize tracer
	clientOTel := otlptracegrpc.NewClient()
	exp, err := otlptrace.New(ctx, clientOTel)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %e", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceVersion("0.0.1"),
		)),
	)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tp, nil
}
