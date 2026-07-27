package opentelemetry

import "errors"

// Config 是 web/opentelemetry 的完整配置，通过 Factory.Build() 构造。
type Config struct {
	ServiceName        string
	Endpoint           string
	URLPath            string
	Insecure           bool
	Headers            map[string]string
	NormalSampler      float64
	LRUMaxSize         int
	ErrorTTLSeconds    int
	BatchTimeout       int
	MaxExportBatchSize int
	Batcher            string
}

// Factory 是 Config 的构建器，在构造时补齐默认值。
type Factory struct {
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
