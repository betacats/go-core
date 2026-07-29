package opentelemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestSpanMarkError(t *testing.T) {
	inner := &spanSliceExporter{}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(inner),
	)
	tracer := tp.Tracer("test")
	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span")
	sm := &TraceSpanMarker{}
	sm.TraceSpanMarkError(ctx)
	span.End()

	if len(inner.spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(inner.spans))
	}
	if inner.spans[0].Status().Code != codes.Error {
		t.Fatalf("expected status Error, got %v", inner.spans[0].Status().Code)
	}
}

func TestSpanMarkOk(t *testing.T) {
	inner := &spanSliceExporter{}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(inner),
	)
	tracer := tp.Tracer("test")
	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span")
	sm := &TraceSpanMarker{}
	sm.TraceSpanMarkOk(ctx)
	span.End()

	if len(inner.spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(inner.spans))
	}
	if inner.spans[0].Status().Code != codes.Ok {
		t.Fatalf("expected status Ok, got %v", inner.spans[0].Status().Code)
	}
}
