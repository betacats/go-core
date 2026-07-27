package opentelemetry

import (
	"testing"
)

// TestConfig_Defaults 验证 NewConfig 默认值。
func TestConfig_Defaults(t *testing.T) {
	cfg := NewConfig("test-svc", "http://localhost:4318", "/v1/traces")

	if cfg.ServiceName != "test-svc" {
		t.Fatalf("expected ServiceName test-svc, got %s", cfg.ServiceName)
	}
	if cfg.Endpoint != "http://localhost:4318" {
		t.Fatalf("expected Endpoint http://localhost:4318, got %s", cfg.Endpoint)
	}
	if cfg.URLPath != "/v1/traces" {
		t.Fatalf("expected URLPath /v1/traces, got %s", cfg.URLPath)
	}
	if cfg.NormalSampler != 0.1 {
		t.Fatalf("expected default NormalSampler 0.1, got %f", cfg.NormalSampler)
	}
	if cfg.LRUMaxSize != 10000 {
		t.Fatalf("expected default LRUMaxSize 10000, got %d", cfg.LRUMaxSize)
	}
	if cfg.ErrorTTLSeconds != 30 {
		t.Fatalf("expected default ErrorTTLSeconds 30, got %d", cfg.ErrorTTLSeconds)
	}
	if cfg.BatchTimeout != 5 {
		t.Fatalf("expected default BatchTimeout 5, got %d", cfg.BatchTimeout)
	}
	if cfg.MaxExportBatchSize != 512 {
		t.Fatalf("expected default MaxExportBatchSize 512, got %d", cfg.MaxExportBatchSize)
	}
	if cfg.Batcher != "batch" {
		t.Fatalf("expected default Batcher batch, got %s", cfg.Batcher)
	}
}

// TestConfig_FieldOverride 验证属性覆盖后默认值不被影响。
func TestConfig_FieldOverride(t *testing.T) {
	cfg := NewConfig("test-svc", "http://localhost:4318", "/custom/traces")
	cfg.NormalSampler = 0.5
	cfg.Insecure = true
	cfg.Headers = map[string]string{"auth": "token"}

	if cfg.URLPath != "/custom/traces" {
		t.Fatalf("expected URLPath /custom/traces, got %s", cfg.URLPath)
	}
	if cfg.NormalSampler != 0.5 {
		t.Fatalf("expected NormalSampler 0.5, got %f", cfg.NormalSampler)
	}
	if !cfg.Insecure {
		t.Fatal("expected Insecure=true")
	}
	if cfg.Headers["auth"] != "token" {
		t.Fatalf("expected Headers[auth]=token, got %s", cfg.Headers["auth"])
	}
	// 未覆盖的字段保持默认值
	if cfg.LRUMaxSize != 10000 {
		t.Fatalf("expected LRUMaxSize 10000, got %d", cfg.LRUMaxSize)
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
