package responsex

import "context"

// Builder 用于构建统一成功/失败响应。
// 这是整个 responsex 的核心对象。
type Builder struct {
	opts Options
}

// New 创建一个新的响应构建器。
// 如果业务方未传 Decoder，则默认启用：
// 1. ResponseCoder 解析能力
// 2. gRPC status error 解析能力
func New(opts Options) *Builder {
	opts = opts.withDefaults()
	if opts.EnableReport && opts.Reporter == nil {
		opts.Reporter = NewSentryReporter()
	}
	if opts.Decoder == nil {
		opts.Decoder = ChainDecoders(
			DecodeResponseCoder(),
			DecodeGRPCStatus(),
		)
	}
	return &Builder{
		opts: opts,
	}
}

// BuildOK 构建统一成功响应。
// 当 data 为 nil 时，会统一输出为空对象，避免前端收到 null。
func (b *Builder) BuildOK(ctx context.Context, data any) Response {
	resp := Response{
		Result: ResultSuccess,
		Code:   b.opts.SuccessCode,
		Msg:    b.opts.SuccessMsg,
		Data:   normalizeData(data),
	}
	b.attachTraceField(ctx, &resp)
	return resp
}

// BuildError 构建统一失败响应。
// 该方法会先解析 err，再套入统一响应结构。
// 如果开启了错误上报，还会在这里触发 Reporter。
func (b *Builder) BuildError(ctx context.Context, err error) Response {
	parsed := b.parseError(ctx, err)

	resp := Response{
		Result: parsed.Result,
		Code:   parsed.Code,
		Msg:    parsed.Msg,
		Data:   b.opts.ErrorData(ctx, parsed),
	}
	b.attachTraceField(ctx, &resp)

	// 只有明确开启上报、注入了上报器且允许上报时，才执行错误上报。
	if b.opts.EnableReport &&
		b.opts.Reporter != nil &&
		b.opts.ShouldReport(ctx, parsed) {
		b.opts.Reporter.Report(ctx, ReporterPayload{
			Response: resp,
			Parsed:   parsed,
		})
	}

	return resp
}

// parseError 将原始 error 解析成统一的 ParsedError。
// 若未命中任何 decoder，则回退到默认错误码和默认错误文案逻辑。
func (b *Builder) parseError(ctx context.Context, err error) ParsedError {
	if err == nil {
		return ParsedError{
			Result: ResultSuccess,
			Code:   b.opts.SuccessCode,
			Msg:    b.opts.SuccessMsg,
		}
	}

	if b.opts.Decoder != nil {
		if parsed, ok := b.opts.Decoder.Decode(ctx, err); ok {
			// 除非 decoder 明确返回成功，否则一律按失败处理。
			if parsed.Result != ResultSuccess {
				parsed.Result = ResultFailure
			}
			if parsed.Code == 0 {
				parsed.Code = b.opts.DefaultErrorCode
			}
			if parsed.Msg == "" {
				parsed.Msg = b.opts.DefaultErrorMsg
			}
			if parsed.Err == nil {
				parsed.Err = err
			}
			return parsed
		}
	}

	// 如果没有命中 decoder，则优先返回 err.Error() 作为 msg。
	// 这样可以尽可能保留原始错误信息。
	msg := err.Error()
	if msg == "" {
		msg = b.opts.DefaultErrorMsg
	}

	return ParsedError{
		Result: ResultFailure,
		Code:   b.opts.DefaultErrorCode,
		Msg:    msg,
		Err:    err,
	}
}

// attachTraceField 按配置为响应附加链路字段。
func (b *Builder) attachTraceField(ctx context.Context, resp *Response) {
	if resp == nil {
		return
	}

	value, ok := b.extractTraceField(ctx)
	if !ok {
		return
	}

	switch b.opts.TraceFieldMode {
	case TraceFieldModeTraceID:
		resp.TraceID = value
	case TraceFieldModeSpan:
		resp.Span = value
	}
}

// extractTraceField 按配置决定是否提取并返回链路字段。
// go-core 不关心你用的是 otel、skywalking 还是别的 tracing 系统。
func (b *Builder) extractTraceField(ctx context.Context) (string, bool) {
	if b.opts.TraceFieldMode == TraceFieldModeNone || b.opts.TraceFieldExtractor == nil {
		return "", false
	}

	value := b.opts.TraceFieldExtractor(ctx)
	if value == "" {
		return "", false
	}

	return value, true
}

// normalizeData 用于统一 nil data 的输出格式。
// 成功响应默认返回空对象，避免前端处理 null。
func normalizeData(data any) any {
	if data == nil {
		return map[string]any{}
	}
	return data
}
