package responsex

import "context"

// ResultCode 表示业务处理结果。
// true 表示成功，false 表示失败。
type ResultCode bool

const (
	// ResultSuccess 表示接口处理成功。
	ResultSuccess ResultCode = true
	// ResultFailure 表示接口处理失败。
	ResultFailure ResultCode = false
)

// Response 是统一的响应体结构。
// 其中 result/code/msg/data 是跨业务系统约定的固定字段。
// 若业务方开启链路字段，则会按配置输出 traceId 或 span。
type Response struct {
	Result  ResultCode `json:"result"`
	Code    int        `json:"code"`
	Msg     string     `json:"msg"`
	Data    any        `json:"data"`
	TraceID string     `json:"traceId,omitempty"`
	Span    string     `json:"span,omitempty"`
}

// ParsedError 表示 error 被解析后的标准结构。
// Builder 会基于该结构生成统一错误响应。
type ParsedError struct {
	Result ResultCode
	Code   int
	Msg    string
	Data   any
	Err    error
}

// ErrorDecoder 用于将任意 error 解析为 ParsedError。
// 业务方可以注入多个 decoder，兼容不同错误模型。
type ErrorDecoder interface {
	Decode(ctx context.Context, err error) (ParsedError, bool)
}

// ErrorDecoderFunc 是 ErrorDecoder 的函数式适配器。
// 方便业务方直接使用函数作为错误解析器。
type ErrorDecoderFunc func(ctx context.Context, err error) (ParsedError, bool)

// Decode 执行函数式错误解析逻辑。
func (f ErrorDecoderFunc) Decode(ctx context.Context, err error) (ParsedError, bool) {
	return f(ctx, err)
}

// ReporterPayload 是错误上报时的统一载荷。
// 外部上报器可基于它完成 sentry、日志平台、告警平台等接入。
type ReporterPayload struct {
	Response Response
	Parsed   ParsedError
}

// Reporter 用于抽象外部错误上报能力。
// go-core 不关心你上报到哪里，只负责定义标准接口。
type Reporter interface {
	Report(ctx context.Context, payload ReporterPayload)
}

// ReporterFunc 是 Reporter 的函数式适配器。
// 方便业务方直接注入函数。
type ReporterFunc func(ctx context.Context, payload ReporterPayload)

// Report 执行函数式上报逻辑。
func (f ReporterFunc) Report(ctx context.Context, payload ReporterPayload) {
	f(ctx, payload)
}

// TraceFieldExtractor 用于从上下文中提取链路追踪标识。
// go-core 不依赖任何 tracing 框架，交给业务方自行注入。
type TraceFieldExtractor func(ctx context.Context) string

// TraceFieldMode 表示响应中链路字段的输出方式。
type TraceFieldMode string

const (
	// TraceFieldModeNone 表示不输出链路字段。
	TraceFieldModeNone TraceFieldMode = ""
	// TraceFieldModeTraceID 表示输出 traceId 字段。
	TraceFieldModeTraceID TraceFieldMode = "traceId"
	// TraceFieldModeSpan 表示输出 span 字段。
	TraceFieldModeSpan TraceFieldMode = "span"
)

// ErrorDataBuilder 用于构建错误响应中的 data 字段。
// 如果业务方不需要特殊 error data，可以不传。
type ErrorDataBuilder func(ctx context.Context, parsed ParsedError) any

// ShouldReportFunc 用于控制某类错误是否需要上报。
// 比如业务错误不报，系统错误才上报。
type ShouldReportFunc func(ctx context.Context, parsed ParsedError) bool

// Options 是 Builder 的全部配置项。
type Options struct {
	// SuccessCode 是成功响应默认 code。
	SuccessCode int
	// SuccessMsg 是成功响应默认 msg。
	SuccessMsg string

	// DefaultErrorCode 是未命中任何错误解析器时的默认错误码。
	DefaultErrorCode int
	// DefaultErrorMsg 是未命中任何错误解析器时的默认错误文案。
	DefaultErrorMsg string

	// TraceFieldMode 控制响应中输出哪一种链路字段。
	// 默认不输出；可选 traceId 或 span。
	TraceFieldMode TraceFieldMode
	// TraceFieldExtractor 用于从 ctx 中提取链路字段值。
	TraceFieldExtractor TraceFieldExtractor

	// EnableReport 控制是否开启错误上报。
	EnableReport bool
	// Reporter 是外部错误上报实现。
	Reporter Reporter
	// ShouldReport 控制某个错误是否需要上报。
	ShouldReport ShouldReportFunc

	// Decoder 用于解析 error。
	Decoder ErrorDecoder
	// ErrorData 用于生成错误响应中的 data 字段。
	ErrorData ErrorDataBuilder
}

// withDefaults 用于为 Options 补齐默认值。
// 这样业务方只配置必要项即可。
func (o Options) withDefaults() Options {
	if o.SuccessMsg == "" {
		o.SuccessMsg = "success"
	}
	if o.DefaultErrorCode == 0 {
		// 这里默认使用 2 作为未知错误码。
		// 与 gRPC Unknown 的语义接近，适合作为兜底值。
		o.DefaultErrorCode = 2
	}
	if o.DefaultErrorMsg == "" {
		o.DefaultErrorMsg = "UNKNOWN"
	}
	if o.ShouldReport == nil {
		o.ShouldReport = func(ctx context.Context, parsed ParsedError) bool {
			return true
		}
	}
	if o.ErrorData == nil {
		o.ErrorData = func(ctx context.Context, parsed ParsedError) any {
			if parsed.Data != nil {
				return parsed.Data
			}
			return map[string]any{}
		}
	}
	return o
}


