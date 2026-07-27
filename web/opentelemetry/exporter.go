package opentelemetry

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// errorAwareExporter 包装一个真实的 SpanExporter，在导出时做 trace 级别的错误感知采样。
//
// 核心逻辑：
//  1. 每个 batch 中检查有哪些 span 被标记为 codes.Error。
//  2. 同一 traceID 中只要有任意一个 span 报错，整条 trace 的所有 span 全部保留。
//  3. 跨 batch 串联：用 LRU 缓存保留最近 ErrorTTLSeconds 秒内的错误 traceID。
//     后续 batch 中同一 trace 的 span 即使没有 Error，也会被全量保留。
//  4. 正常 trace（无 Error）的每个 span 按 NormalSampler 独立随机采样。
type errorAwareExporter struct {
	// inner 被包装的真实 SpanExporter（OTLP HTTP exporter）。
	inner sdktrace.SpanExporter
	// cfg 配置参数，包含采样率、LRU 容量等。
	cfg Config
	// lru 跨 batch 错误 traceID 缓存。
	lru *traceLRU
	// mu 保护 lru 并发访问的互斥锁。
	mu sync.Mutex
}

func newErrorAwareExporter(inner sdktrace.SpanExporter, cfg Config) *errorAwareExporter {
	return &errorAwareExporter{
		inner: inner,
		cfg:   cfg,
		lru:   newTraceLRU(cfg.LRUMaxSize, cfg.ErrorTTLSeconds),
	}
}

// ExportSpans 执行 trace 级别的采样过滤，然后将保留的 span 委托给内部 exporter。
func (e *errorAwareExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if len(spans) == 0 {
		return nil
	}

	// Phase 1: 扫描本 batch 中有 Error 的 traceID
	errorTraces := make(map[trace.TraceID]struct{}, len(spans))
	for _, s := range spans {
		if s.Status().Code == codes.Error {
			errorTraces[s.SpanContext().TraceID()] = struct{}{}
		}
	}

	// Phase 2: 判断哪些 trace 需要全量保留
	infected := make(map[trace.TraceID]struct{}, len(errorTraces))
	now := time.Now()

	e.mu.Lock()
	for _, s := range spans {
		tid := s.SpanContext().TraceID()
		if _, ok := errorTraces[tid]; ok {
			infected[tid] = struct{}{}
		} else if e.lru.contains(tid, now) {
			infected[tid] = struct{}{}
		}
	}

	// Phase 3: 更新 LRU — 将本次有 Error 的 traceID 写入缓存
	for tid := range errorTraces {
		e.lru.add(tid, now)
	}
	e.mu.Unlock()

	// Phase 4: 过滤 span
	kept := make([]sdktrace.ReadOnlySpan, 0, len(spans))
	for _, s := range spans {
		tid := s.SpanContext().TraceID()
		if _, ok := infected[tid]; ok {
			kept = append(kept, s)
		} else if sample(e.cfg.NormalSampler) {
			kept = append(kept, s)
		}
	}

	if len(kept) == 0 {
		return nil
	}

	return e.inner.ExportSpans(ctx, kept)
}

func (e *errorAwareExporter) Shutdown(ctx context.Context) error {
	return e.inner.Shutdown(ctx)
}

// traceLRU 是一个带有 TTL 过期能力的 LRU 缓存，存储"有 Error 的 traceID"。
// 用于跨 batch 串联同一 trace 的 span。
type traceLRU struct {
	// maxSize 缓存最大条目数，超过时淘汰最旧元素。
	maxSize int
	// ttl 每条错误 traceID 的保留时长。
	ttl time.Duration
	// entries traceID 到链表元素的快速查找映射。
	entries map[trace.TraceID]*lruElement
	// list 按添加顺序排列的元素列表，头旧尾新。
	list []*lruElement
}

type lruElement struct {
	// tid 错误 trace 的 ID。
	tid trace.TraceID
	// expiresAt 该条目的过期时间，超过后不再影响采样决策。
	expiresAt time.Time
}

func newTraceLRU(maxSize int, ttlSeconds int) *traceLRU {
	return &traceLRU{
		maxSize: maxSize,
		ttl:     time.Duration(ttlSeconds) * time.Second,
		entries: make(map[trace.TraceID]*lruElement),
		list:    make([]*lruElement, 0, maxSize),
	}
}

func (l *traceLRU) contains(tid trace.TraceID, now time.Time) bool {
	el, ok := l.entries[tid]
	if !ok {
		return false
	}
	if !now.Before(el.expiresAt) {
		l.remove(el)
		return false
	}
	return true
}

func (l *traceLRU) add(tid trace.TraceID, now time.Time) {
	if _, ok := l.entries[tid]; ok {
		return
	}

	el := &lruElement{
		tid:       tid,
		expiresAt: now.Add(l.ttl),
	}

	if len(l.list) >= l.maxSize {
		l.evict()
	}

	l.entries[tid] = el
	l.list = append(l.list, el)
}

func (l *traceLRU) remove(el *lruElement) {
	delete(l.entries, el.tid)

	for i, e := range l.list {
		if e == el {
			l.list = append(l.list[:i], l.list[i+1:]...)
			return
		}
	}
}

func (l *traceLRU) evict() {
	if len(l.list) == 0 {
		return
	}

	oldest := l.list[0]
	delete(l.entries, oldest.tid)
	l.list = l.list[1:]
}
