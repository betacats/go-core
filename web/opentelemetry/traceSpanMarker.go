package opentelemetry

import (
	"context"

	"github.com/betacats/go-core/web/responsex"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type TraceSpanMarker struct {
	resp responsex.Response
}

// 将当前Span标记为error
func (s *TraceSpanMarker) TraceSpanMarkError(ctx context.Context) {
	trace.SpanFromContext(ctx).SetStatus(codes.Error, "Span Error")
}

// 将当前Span标记为success
func (s *TraceSpanMarker) TraceSpanMarkOk(ctx context.Context, resp responsex.Response) {
	trace.SpanFromContext(ctx).SetStatus(codes.Ok, "Span Ok")
}

func NewTraceSpanMarker() *TraceSpanMarker {
	return &TraceSpanMarker{}
}
