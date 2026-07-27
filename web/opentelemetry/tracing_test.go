package opentelemetry

import (
	"testing"
)

// TestFactory_Defaults 验证 Factory 默认值。
func TestFactory_Defaults(t *testing.T) {
	f := NewFactory()

	if f.config.URLPath != "/v1/traces" {
		t.Fatalf("expected default URLPath /v1/traces, got %s", f.config.URLPath)
	}
	if f.config.NormalSampler != 0.1 {
		t.Fatalf("expected default NormalSampler 0.1, got %f", f.config.NormalSampler)
	}
	if f.config.LRUMaxSize != 10000 {
		t.Fatalf("expected default LRUMaxSize 10000, got %d", f.config.LRUMaxSize)
	}
	if f.config.ErrorTTLSeconds != 30 {
		t.Fatalf("expected default ErrorTTLSeconds 30, got %d", f.config.ErrorTTLSeconds)
	}
	if f.config.BatchTimeout != 5 {
		t.Fatalf("expected default BatchTimeout 5, got %d", f.config.BatchTimeout)
	}
	if f.config.MaxExportBatchSize != 512 {
		t.Fatalf("expected default MaxExportBatchSize 512, got %d", f.config.MaxExportBatchSize)
	}
	if f.config.Batcher != "batch" {
		t.Fatalf("expected default Batcher batch, got %s", f.config.Batcher)
	}
}

// TestFactory_CustomValues 验证自定义值不会被默认值覆盖。
func TestFactory_CustomValues(t *testing.T) {
	f := NewFactory().
		WithServiceName("test-svc").
		WithEndpoint("http://localhost:4318").
		WithURLPath("/custom/traces").
		WithNormalSampler(0.5).
		WithLRUMaxSize(5000).
		WithErrorTTLSeconds(60).
		WithBatchTimeout(10).
		WithMaxExportBatchSize(1024)

	if f.config.URLPath != "/custom/traces" {
		t.Fatalf("expected URLPath /custom/traces, got %s", f.config.URLPath)
	}
	if f.config.NormalSampler != 0.5 {
		t.Fatalf("expected NormalSampler 0.5, got %f", f.config.NormalSampler)
	}
	if f.config.LRUMaxSize != 5000 {
		t.Fatalf("expected LRUMaxSize 5000, got %d", f.config.LRUMaxSize)
	}
	if f.config.ErrorTTLSeconds != 60 {
		t.Fatalf("expected ErrorTTLSeconds 60, got %d", f.config.ErrorTTLSeconds)
	}
	if f.config.BatchTimeout != 10 {
		t.Fatalf("expected BatchTimeout 10, got %d", f.config.BatchTimeout)
	}
	if f.config.MaxExportBatchSize != 1024 {
		t.Fatalf("expected MaxExportBatchSize 1024, got %d", f.config.MaxExportBatchSize)
	}
}

// TestFactory_Build_Valid 验证 Build 成功。
func TestFactory_Build_Valid(t *testing.T) {
	cfg, err := NewFactory().
		WithServiceName("test-svc").
		WithEndpoint("http://localhost:4318").
		Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if cfg.ServiceName != "test-svc" {
		t.Fatalf("expected ServiceName test-svc, got %s", cfg.ServiceName)
	}
	if cfg.Endpoint != "http://localhost:4318" {
		t.Fatalf("expected Endpoint http://localhost:4318, got %s", cfg.Endpoint)
	}
}

// TestFactory_Build_MissingRequired 验证缺少必填字段时 Build 返回错误。
func TestFactory_Build_MissingRequired(t *testing.T) {
	_, err := NewFactory().Build()
	if err == nil {
		t.Fatal("expected error for missing required fields")
	}

	_, err = NewFactory().WithServiceName("svc").Build()
	if err == nil {
		t.Fatal("expected error for missing Endpoint")
	}

	_, err = NewFactory().WithEndpoint("http://localhost:4318").Build()
	if err == nil {
		t.Fatal("expected error for missing ServiceName")
	}
}

// TestSample_Bounds 验证采样边界行为。
func TestSample_Bounds(t *testing.T) {
	if sample(0) {
		t.Fatal("expected sample(0) to be false")
	}
	if !sample(1) {
		t.Fatal("expected sample(1) to be true")
	}
	if sample(-0.1) {
		t.Fatal("expected sample(-0.1) to be false")
	}
}
