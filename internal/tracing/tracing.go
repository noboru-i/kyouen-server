package tracing

import (
	"context"
	"log"

	cloudtrace "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Init initializes the OpenTelemetry TracerProvider with Cloud Trace exporter.
// Returns a shutdown function that must be called before the process exits.
func Init(ctx context.Context, projectID string) func() {
	exporter, err := cloudtrace.New(cloudtrace.WithProjectID(projectID))
	if err != nil {
		log.Printf("Warning: failed to create Cloud Trace exporter: %v (tracing disabled)", err)
		return func() {}
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}
}
