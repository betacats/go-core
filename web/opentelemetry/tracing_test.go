package opentelemetry

import (
	"testing"
)

// TestConfigWithDefaults 验证 Config 默认值补齐逻辑。
func TestConfigWithDefaults(t *testing.T) {
	cfg := Config{}.withDefaults()

	if cfg.URLPath != "/v1/traces" {
		t.Fatalf("expected default URLPath /v1/traces, got %s", cfg.URLPath)
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
}

// TestConfigWithDefaults_CustomValues 验证自定义值不会被覆盖。
func TestConfigWithDefaults_CustomValues(t *testing.T) {
	cfg := Config{
		URLPath:           "/custom/traces",
		NormalSampler:     0.5,
		LRUMaxSize:        5000,
		ErrorTTLSeconds:   60,
		BatchTimeout:      10,
		MaxExportBatchSize: 1024,
	}.withDefaults()

	if cfg.URLPath != "/custom/traces" {
		t.Fatalf("expected URLPath /custom/traces, got %s", cfg.URLPath)
	}
	if cfg.NormalSampler != 0.5 {
		t.Fatalf("expected NormalSampler 0.5, got %f", cfg.NormalSampler)
	}
	if cfg.LRUMaxSize != 5000 {
		t.Fatalf("expected LRUMaxSize 5000, got %d", cfg.LRUMaxSize)
	}
	if cfg.ErrorTTLSeconds != 60 {
		t.Fatalf("expected ErrorTTLSeconds 60, got %d", cfg.ErrorTTLSeconds)
	}
	if cfg.BatchTimeout != 10 {
		t.Fatalf("expected BatchTimeout 10, got %d", cfg.BatchTimeout)
	}
	if cfg.MaxExportBatchSize != 1024 {
		t.Fatalf("expected MaxExportBatchSize 1024, got %d", cfg.MaxExportBatchSize)
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
