package otel_instrumentation

import (
	"encoding/json"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Utility function to add an event to the span with some metadata
// This can be helpful in some situations but not the best practice
func AddLogEvent(span trace.Span, metadata any) error {
	mJson, err := json.Marshal(metadata)
	if err == nil && len(mJson) != 0 {
		attrs := make([]attribute.KeyValue, 0)
		logMessageKey := attribute.Key("log.metadata")
		attrs = append(attrs, logMessageKey.String(string(mJson)))
		span.AddEvent("log", trace.WithAttributes(attrs...))
	}

	return err
}
