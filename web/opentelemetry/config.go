package opentelemetry

// NewConfig 创建 Config，注入全部默认值。
func NewConfig(serviceName, endpoint, urlPath string) Config {
	return Config{
		ServiceName:        serviceName,
		Endpoint:           endpoint,
		URLPath:            urlPath,
		NormalSampler:      0.1,
		LRUMaxSize:         10000,
		ErrorTTLSeconds:    30,
		BatchTimeout:       5,
		MaxExportBatchSize: 512,
		Batcher:            "otlphttp",
	}
}

type Config struct {
	// ServiceName 服务名称，会作为资源属性写入 span。
	ServiceName string
	// Endpoint OTLP collector 地址，例如 "http://localhost:4318"。
	Endpoint string
	// URLPath OTLP traces 路径，例如 "/v1/traces"。
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

// WithNormalSampler 设置正常 span 采样率 [0,1]。
func (c *Config) WithNormalSampler(v float64) *Config { c.NormalSampler = v; return c }

// WithLRUMaxSize 设置错误 traceID 缓存最大条目数。
func (c *Config) WithLRUMaxSize(v int) *Config { c.LRUMaxSize = v; return c }

// WithErrorTTLSeconds 设置错误 traceID 缓存保留秒数。
func (c *Config) WithErrorTTLSeconds(v int) *Config { c.ErrorTTLSeconds = v; return c }

// WithBatchTimeout 设置批量导出最大等待秒数。
func (c *Config) WithBatchTimeout(v int) *Config { c.BatchTimeout = v; return c }

// WithMaxExportBatchSize 设置单次批量导出最大 span 数。
func (c *Config) WithMaxExportBatchSize(v int) *Config { c.MaxExportBatchSize = v; return c }

// WithBatcher 设置 span 处理器类型（当前保留供后续扩展）。
func (c *Config) WithBatcher(v string) *Config { c.Batcher = v; return c }
