package opentelemetry

// Config 是 web/opentelemetry 的配置项。
type Config struct {
	// ServiceName 是服务名称，会作为资源属性写入 span。
	ServiceName string

	// Endpoint 是 OTLP collector 的地址。
	// 例如: "http://localhost:4318"（HTTP）或 "localhost:4317"（gRPC）。
	// URL 内置的认证信息（如 "http://key@host"）也会被保留。
	Endpoint string

	// URLPath 指定 OTLP traces 的路径，默认 "/v1/traces"。
	URLPath string

	// Insecure 为 true 时跳过 TLS 证书校验。
	Insecure bool

	// Headers 是额外附加到每个 OTLP 请求的 HTTP 头部。
	// 可用于 Bearer Token、API Key 等鉴权方式。
	Headers map[string]string

	// NormalSampler 是正常（无错误）span 的采样率，取值范围 [0, 1]。
	// 默认 0.1 表示 10%。
	NormalSampler float64

	// LRUMaxSize 是跨 batch 错误 traceID 缓存的最大条目数。
	// 默认 10000。
	LRUMaxSize int

	// ErrorTTLSeconds 是错误 traceID 在缓存中的保留时间（秒）。
	// 在该时间内，同一 trace 的所有 span 都会 100% 上报。
	// 默认 30。
	ErrorTTLSeconds int

	// BatchTimeout 是 span 批量导出的最大等待时间，默认 5s。
	BatchTimeout int

	// MaxExportBatchSize 是单次批量导出的最大 span 数，默认 512。
	MaxExportBatchSize int
}

func (c Config) withDefaults() Config {
	if c.URLPath == "" {
		c.URLPath = "/v1/traces"
	}
	if c.NormalSampler <= 0 || c.NormalSampler > 1 {
		c.NormalSampler = 0.1
	}
	if c.LRUMaxSize <= 0 {
		c.LRUMaxSize = 10000
	}
	if c.ErrorTTLSeconds <= 0 {
		c.ErrorTTLSeconds = 30
	}
	if c.BatchTimeout <= 0 {
		c.BatchTimeout = 5
	}
	if c.MaxExportBatchSize <= 0 {
		c.MaxExportBatchSize = 512
	}
	return c
}
