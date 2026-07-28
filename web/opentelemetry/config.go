package opentelemetry

type Config struct {
	// ServiceName 服务名称，会作为资源属性写入 span。
	ServiceName string
	// Endpoint OTLP collector 地址，例如 "http://localhost:4318"。
	Endpoint string
	// URLPath OTLP traces 路径，例如 "/v1/traces"。
	URLPath string

	// Secure 为 true 时开启 TLS 证书校验。
	Secure bool
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

func (c *Config) GetServiceName() string { return c.ServiceName }
func (c *Config) GetEndpoint() string    { return c.Endpoint }

func (c *Config) GetURLPath() string {
	if c.URLPath != "" {
		return c.URLPath
	}
	return "/v1/traces"
}

func (c *Config) GetSecure() bool               { return c.Secure }
func (c *Config) GetHeaders() map[string]string { return c.Headers }

func (c *Config) GetNormalSampler() float64 {
	if c.NormalSampler <= 0 || c.NormalSampler > 1 {
		return 0.1
	}
	return c.NormalSampler
}

func (c *Config) GetLRUMaxSize() int {
	if c.LRUMaxSize <= 0 {
		return 10000
	}
	return c.LRUMaxSize
}

func (c *Config) GetErrorTTLSeconds() int {
	if c.ErrorTTLSeconds <= 0 {
		return 30
	}
	return c.ErrorTTLSeconds
}

func (c *Config) GetBatchTimeout() int {
	if c.BatchTimeout <= 0 {
		return 5
	}
	return c.BatchTimeout
}

func (c *Config) GetMaxExportBatchSize() int {
	if c.MaxExportBatchSize <= 0 {
		return 512
	}
	return c.MaxExportBatchSize
}

func (c *Config) GetBatcher() string {
	if c.Batcher == "" {
		return "batch"
	}
	return c.Batcher
}
