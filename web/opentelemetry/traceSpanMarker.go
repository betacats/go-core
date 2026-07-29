package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type TraceSpanMarker struct {
}

// 将当前Span标记为error
func (s *TraceSpanMarker) TraceSpanMarkError(ctx context.Context) {
	trace.SpanFromContext(ctx).SetStatus(codes.Error, "Span Error")
}

// 将当前Span标记为success
func (s *TraceSpanMarker) TraceSpanMarkOk(ctx context.Context) {
	trace.SpanFromContext(ctx).SetStatus(codes.Ok, "Span Ok")
}

func NewTraceSpanMarker() *TraceSpanMarker {
	return &TraceSpanMarker{}
}
