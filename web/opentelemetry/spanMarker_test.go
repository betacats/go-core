package opentelemetry

import (
	"context"
	"testing"

	"github.com/betacats/go-core/web/responsex"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestMarkSpanError(t *testing.T) {
	inner := &spanSliceExporter{}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(inner),
	)
	tracer := tp.Tracer("test")
	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span")
	resp := responsex.Response{Result: responsex.ResultFailure, Code: 7, Msg: "error"}
	MarkSpanError(ctx, resp)
	span.End()

	if len(inner.spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(inner.spans))
	}
	if inner.spans[0].Status().Code != codes.Error {
		t.Fatalf("expected status Error, got %v", inner.spans[0].Status().Code)
	}
}

func TestMarkSpanOk(t *testing.T) {
	inner := &spanSliceExporter{}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(inner),
	)
	tracer := tp.Tracer("test")
	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span")
	resp := responsex.Response{Result: responsex.ResultSuccess, Code: 0, Msg: "success"}
	MarkSpanOk(ctx, resp)
	span.End()

	if len(inner.spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(inner.spans))
	}
	if inner.spans[0].Status().Code != codes.Ok {
		t.Fatalf("expected status Ok, got %v", inner.spans[0].Status().Code)
	}
}
