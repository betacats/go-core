package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const defaultAttrValueMaxLen = 4096

// SetSpanAttr injects a key-value attribute into the span from ctx.
// Value is truncated to 4096 bytes to prevent oversized attributes.
// No-op if there's no valid span in context.
func SetSpanAttr(ctx context.Context, key, value string) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}
	if len(value) > defaultAttrValueMaxLen {
		value = value[:defaultAttrValueMaxLen]
	}
	span.SetAttributes(attribute.String(key, value))
}
