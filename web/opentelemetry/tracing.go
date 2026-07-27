package opentelemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
)

// InitTracing 初始化全局 TracerProvider。
//
// 它会创建一个 AlwaysSample 的 TracerProvider，确保每个请求都产生 recording span，
// 然后通过 errorAwareExporter 在导出阶段做 trace 级别的智能过滤。
//
// 在 go-zero 项目中，调用此函数前应先将 Telemetry.Disabled 设为 true，
func InitTracing(cfg Config) (shutdown func(context.Context) error, err error) {
	exp, err := newOTLPExporter(cfg)
	if err != nil {
		return nil, fmt.Errorf("opentelemetry: create otlp exporter: %w", err)
	}

	awareExp := newErrorAwareExporter(exp, cfg)
	processor := newSpanProcessor(cfg, awareExp)

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("opentelemetry: create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(processor),
	)

	otel.SetTracerProvider(tp)

	globalShutdown = func(ctx context.Context) error {
		return errors.Join(tp.Shutdown(ctx), exp.Shutdown(ctx))
	}

	return globalShutdown, nil
}

// newSpanProcessor 根据 cfg.Batcher 创建对应的 SpanProcessor。
func newSpanProcessor(cfg Config, exp sdktrace.SpanExporter) sdktrace.SpanProcessor {
	return sdktrace.NewBatchSpanProcessor(exp,
		sdktrace.WithBatchTimeout(time.Duration(cfg.BatchTimeout)*time.Second),
		sdktrace.WithMaxExportBatchSize(cfg.MaxExportBatchSize),
	)
}

// newOTLPExporter 创建 OTLP HTTP exporter。
func newOTLPExporter(cfg Config) (*otlptrace.Exporter, error) {
	ctx := context.Background()

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithURLPath(cfg.URLPath),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
	}

	client := otlptracehttp.NewClient(opts...)
	return otlptrace.New(ctx, client)
}

var globalShutdown func(context.Context) error

// StopTracing 优雅关闭全局 TracerProvider。
// 该函数期望在服务退出时被调用（如 proc.AddShutdownListener）。
func StopTracing() {
	if globalShutdown == nil {
		return
	}
	_ = globalShutdown(context.Background())
}
