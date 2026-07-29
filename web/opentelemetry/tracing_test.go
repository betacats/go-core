package opentelemetry

import (
	"testing"
)

// TestConfig_Defaults 验证 getter 默认值。
func TestConfig_Defaults(t *testing.T) {
	cfg := Config{
		ServiceName: "test-svc",
		Endpoint:    "http://localhost:4318",
		URLPath:     "/v1/traces",
	}

	if cfg.GetNormalSampler() != 0.1 {
		t.Fatalf("expected default NormalSampler 0.1, got %f", cfg.GetNormalSampler())
	}
	if cfg.GetLRUMaxSize() != 10000 {
		t.Fatalf("expected default LRUMaxSize 10000, got %d", cfg.GetLRUMaxSize())
	}
	if cfg.GetErrorTTLSeconds() != 30 {
		t.Fatalf("expected default ErrorTTLSeconds 30, got %d", cfg.GetErrorTTLSeconds())
	}
	if cfg.GetBatchTimeout() != 5 {
		t.Fatalf("expected default BatchTimeout 5, got %d", cfg.GetBatchTimeout())
	}
	if cfg.GetMaxExportBatchSize() != 512 {
		t.Fatalf("expected default MaxExportBatchSize 512, got %d", cfg.GetMaxExportBatchSize())
	}
	if cfg.GetBatcher() != "batch" {
		t.Fatalf("expected default Batcher batch, got %s", cfg.GetBatcher())
	}
	if cfg.GetURLPath() != "/v1/traces" {
		t.Fatalf("expected URLPath /v1/traces, got %s", cfg.GetURLPath())
	}
}

// TestConfig_FieldOverride 验证属性覆盖后 getter 返回自定义值。
func TestConfig_FieldOverride(t *testing.T) {
	cfg := Config{
		ServiceName:   "test-svc",
		Endpoint:      "http://localhost:4318",
		URLPath:       "/custom/traces",
		NormalSampler: 0.5,
		Secure:        true,
		Headers:       map[string]string{"auth": "token"},
	}

	if cfg.GetURLPath() != "/custom/traces" {
		t.Fatalf("expected URLPath /custom/traces, got %s", cfg.GetURLPath())
	}
	if cfg.GetNormalSampler() != 0.5 {
		t.Fatalf("expected NormalSampler 0.5, got %f", cfg.GetNormalSampler())
	}
	if !cfg.GetSecure() {
		t.Fatal("expected Insecure=true")
	}
	if cfg.GetHeaders()["auth"] != "token" {
		t.Fatalf("expected Headers[auth]=token, got %s", cfg.GetHeaders()["auth"])
	}
	// 未覆盖的字段 getter 返回默认值
	if cfg.GetLRUMaxSize() != 10000 {
		t.Fatalf("expected LRUMaxSize 10000, got %d", cfg.GetLRUMaxSize())
	}
}

// TestConfig_EmptyDefaults 验证零值 getter 返回默认值。
func TestConfig_EmptyDefaults(t *testing.T) {
	cfg := Config{
		ServiceName: "test-svc",
		Endpoint:    "http://localhost:4318",
		URLPath:     "",
	}

	if cfg.GetURLPath() != "/v1/traces" {
		t.Fatalf("expected default URLPath /v1/traces, got %s", cfg.GetURLPath())
	}
	if cfg.GetNormalSampler() != 0.1 {
		t.Fatalf("expected default NormalSampler 0.1, got %f", cfg.GetNormalSampler())
	}
	if cfg.GetBatcher() != "batch" {
		t.Fatalf("expected default Batcher batch, got %s", cfg.GetBatcher())
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
