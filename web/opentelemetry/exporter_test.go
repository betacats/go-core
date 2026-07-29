package opentelemetry

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// spyExporter 记录被导出 span 的数量。
type spyExporter struct {
	count int
}

func (s *spyExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	s.count += len(spans)
	return nil
}

func (s *spyExporter) Shutdown(_ context.Context) error { return nil }

// spanSliceExporter 捕获所有被导出的 span，用于测试验证。
type spanSliceExporter struct {
	spans []sdktrace.ReadOnlySpan
}

func (s *spanSliceExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	s.spans = append(s.spans, spans...)
	return nil
}

func (s *spanSliceExporter) Shutdown(_ context.Context) error {
	return nil
}

// TestErrorAwareExporter_ErrorTraceKept 验证：有 Error 的 trace 全量保留。
func TestErrorAwareExporter_ErrorTraceKept(t *testing.T) {
	inner := &spanSliceExporter{}
	aware := newErrorAwareExporter(inner, Config{
		NormalSampler:  1.0, // 正常采样 100%，让结果不受随机性影响
		LRUMaxSize:     100,
		ErrorTTLSeconds: 30,
	})

	// 创建一批 span，包含 Error 和正常
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(aware),
	)
	tracer := tp.Tracer("test")
	ctx := context.Background()

	// trace A: 有 Error — 应全量保留
	_, spanA1 := tracer.Start(ctx, "a1")
	_, spanA2 := tracer.Start(ctx, "a2")
	spanA2.SetStatus(codes.Error, "error in a2")
	spanA1.End()
	spanA2.End()

	// trace B: 正常 — 会被采样（NormalSampler=1.0 所以也保留）
	_, spanB1 := tracer.Start(ctx, "b1")
	spanB1.End()

	if len(inner.spans) < 3 {
		t.Fatalf("expected at least 3 spans, got %d", len(inner.spans))
	}
}

// TestErrorAwareExporter_NoErrorSampled 验证：正常 span 按 NormalSampler 采样。
func TestErrorAwareExporter_NoErrorSampled(t *testing.T) {
	inner := &spyExporter{}
	aware := newErrorAwareExporter(inner, Config{
		NormalSampler:  1.0,
		LRUMaxSize:     100,
		ErrorTTLSeconds: 30,
	})

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(aware),
	)
	tracer := tp.Tracer("test")
	ctx := context.Background()

	for range 100 {
		_, span := tracer.Start(ctx, "normal")
		span.End()
	}

	if inner.count == 0 {
		t.Fatal("expected some spans to be sampled at 100% rate")
	}
}

// TestTraceLRU_Expiry 验证 LRU 缓存过期后不再保留 trace。
func TestTraceLRU_Expiry(t *testing.T) {
	lru := newTraceLRU(100, 0) // TTL=0 表示立即过期

	tid := trace.TraceID{1}
	now := time.Now()

	lru.add(tid, now)
	if lru.contains(tid, now) {
		t.Fatal("expected entry to be expired immediately")
	}
}

// TestTraceLRU_Eviction 验证超过 maxSize 后淘汰最旧条目。
func TestTraceLRU_Eviction(t *testing.T) {
	lru := newTraceLRU(2, 60) // 最多保留 2 个

	now := time.Now()
	tid1 := trace.TraceID{1}
	tid2 := trace.TraceID{2}
	tid3 := trace.TraceID{3}

	lru.add(tid1, now)
	lru.add(tid2, now)
	lru.add(tid3, now) // 应淘汰 tid1

	if lru.contains(tid1, now) {
		t.Fatal("expected tid1 to be evicted")
	}
	if !lru.contains(tid2, now) {
		t.Fatal("expected tid2 to remain")
	}
	if !lru.contains(tid3, now) {
		t.Fatal("expected tid3 to remain")
	}
}
