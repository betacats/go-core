package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SetSpanAttr injects a key-value attribute into the span from ctx.
// No-op if there's no valid span in context.
func SetSpanAttr(ctx context.Context, key, value string) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}
	span.SetAttributes(attribute.String(key, value))
}
