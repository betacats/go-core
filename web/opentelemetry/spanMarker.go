package opentelemetry

import (
	"context"

	"github.com/betacats/go-core/web/responsex"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type SpanMarker struct {
	resp responsex.Response
}

// 将当前Span标记为error
func (s *SpanMarker) SpanMarkError(ctx context.Context) {
	trace.SpanFromContext(ctx).SetStatus(codes.Error, "Span Error")
}

// 将当前Span标记为success
func (s *SpanMarker) SpanMarkOk(ctx context.Context, resp responsex.Response) {
	trace.SpanFromContext(ctx).SetStatus(codes.Ok, "Span Ok")
}
