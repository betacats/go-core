package opentelemetry

import "errors"

// Config 是 web/opentelemetry 的完整配置，通过 Factory.Build() 构造。
type Config struct {
	// ServiceName 服务名称，会作为资源属性写入 span。
	ServiceName string
	// Endpoint OTLP collector 地址，例如 "http://localhost:4318"。
	Endpoint string
	// URLPath OTLP traces 路径，默认 "/v1/traces"。
	URLPath string
	// Insecure 为 true 时跳过 TLS 证书校验。
	Insecure bool
	// Headers 附加到每个 OTLP 请求的 HTTP 头部，可用于 ARMS 认证。
	Headers map[string]string
	// NormalSampler 正常 span 采样率 [0,1]，默认 0.1。
	NormalSampler float64
	// LRUMaxSize 错误 traceID 缓存最大条目数，默认 10000。
	LRUMaxSize int
	// ErrorTTLSeconds 错误 traceID 缓存保留秒数，默认 30。
	ErrorTTLSeconds int
	// BatchTimeout 批量导出最大等待秒数，默认 5。
	BatchTimeout int
	// MaxExportBatchSize 单次批量导出最大 span 数，默认 512。
	MaxExportBatchSize int
	// Batcher span 处理器类型，保留供后续扩展（当前固定为 batch）。
	Batcher string
}

// Factory 是 Config 的构建器，在构造时补齐默认值。
type Factory struct {
	// config 正在构建的配置对象。
	config Config
}

// NewFactory 创建一个已注入默认值的 Factory。
func NewFactory() *Factory {
	return &Factory{
		config: Config{
			URLPath:            "/v1/traces",
			NormalSampler:      0.1,
			LRUMaxSize:         10000,
			ErrorTTLSeconds:    30,
			BatchTimeout:       5,
			MaxExportBatchSize: 512,
			Batcher:            "batch",
		},
	}
}

func (f *Factory) WithServiceName(v string) *Factory        { f.config.ServiceName = v; return f }
func (f *Factory) WithEndpoint(v string) *Factory           { f.config.Endpoint = v; return f }
func (f *Factory) WithURLPath(v string) *Factory            { f.config.URLPath = v; return f }
func (f *Factory) WithInsecure(v bool) *Factory             { f.config.Insecure = v; return f }
func (f *Factory) WithHeaders(v map[string]string) *Factory { f.config.Headers = v; return f }
func (f *Factory) WithNormalSampler(v float64) *Factory     { f.config.NormalSampler = v; return f }
func (f *Factory) WithLRUMaxSize(v int) *Factory            { f.config.LRUMaxSize = v; return f }
func (f *Factory) WithErrorTTLSeconds(v int) *Factory       { f.config.ErrorTTLSeconds = v; return f }
func (f *Factory) WithBatchTimeout(v int) *Factory          { f.config.BatchTimeout = v; return f }
func (f *Factory) WithMaxExportBatchSize(v int) *Factory    { f.config.MaxExportBatchSize = v; return f }
func (f *Factory) WithBatcher(v string) *Factory            { f.config.Batcher = v; return f }

// Build 返回补齐默认值后的 Config，并校验必填字段。
func (f *Factory) Build() (Config, error) {
	if f.config.ServiceName == "" {
		return Config{}, errors.New("opentelemetry: ServiceName is required")
	}
	if f.config.Endpoint == "" {
		return Config{}, errors.New("opentelemetry: Endpoint is required")
	}
	return f.config, nil
}
