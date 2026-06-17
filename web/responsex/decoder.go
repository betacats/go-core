package responsex

import (
	"context"
	stdErrors "errors"

	grpcstatus "google.golang.org/grpc/status"
)

type grpcStatusProvider interface {
	GRPCStatus() *grpcstatus.Status
}

// ResponseCoder 表示可直接提供响应 code 和 msg 的错误类型。
// 业务项目只要让自定义错误实现这个接口，就能被统一解析。
// 例如 web/errorx.CodeError 只要实现了 ResponseCode/ResponseMsg，
// Builder.BuildError 就能在默认配置下直接识别它，无需业务方额外注入 decoder。
type ResponseCoder interface {
	ResponseCode() int
	ResponseMsg() string
}

// ResponseResultProvider 表示错误对象可额外提供业务 result。
// 若未实现，则默认按失败处理。
type ResponseResultProvider interface {
	ResponseResult() ResultCode
}

// ResponseDataProvider 表示错误对象可额外提供响应 data。
// 若未实现，则 error response 的 data 使用默认逻辑。
type ResponseDataProvider interface {
	ResponseData() any
}

// ChainDecoders 将多个错误解析器串联起来。
// 解析时按顺序执行，谁先命中谁生效。
func ChainDecoders(decoders ...ErrorDecoder) ErrorDecoder {
	return ErrorDecoderFunc(func(ctx context.Context, err error) (ParsedError, bool) {
		for _, decoder := range decoders {
			if decoder == nil {
				continue
			}
			if parsed, ok := decoder.Decode(ctx, err); ok {
				return parsed, true
			}
		}
		return ParsedError{}, false
	})
}

// DecodeResponseCoder 用于解析实现了 ResponseCoder 的错误。
// 这是最通用的一种业务错误接入方式。
// 解析时会用 errors.As 判断 err 是否“长得像” ResponseCoder：
// 只要错误对象实现了 ResponseCode() 和 ResponseMsg()，就会命中。
func DecodeResponseCoder() ErrorDecoder {
	return ErrorDecoderFunc(func(ctx context.Context, err error) (ParsedError, bool) {
		var coder ResponseCoder
		if !stdErrors.As(err, &coder) {
			return ParsedError{}, false
		}

		parsed := ParsedError{
			Result: ResultFailure,
			Code:   coder.ResponseCode(),
			Msg:    coder.ResponseMsg(),
			Err:    err,
		}

		// 如果错误对象自己能表达 Result，则优先使用其结果。
		var resultProvider ResponseResultProvider
		if stdErrors.As(err, &resultProvider) {
			parsed.Result = resultProvider.ResponseResult()
		}

		// 如果错误对象自己能表达 Data，则将其透传到响应中。
		var dataProvider ResponseDataProvider
		if stdErrors.As(err, &dataProvider) {
			parsed.Data = dataProvider.ResponseData()
		}

		return parsed, true
	})
}

// DecodeGRPCStatus 用于解析 gRPC / RPC status 错误。
// 只有当 err 能被识别为标准 gRPC status error 时才会命中，
// 普通 error 不会误判到这里。
func DecodeGRPCStatus() ErrorDecoder {
	return ErrorDecoderFunc(func(ctx context.Context, err error) (ParsedError, bool) {
		var provider grpcStatusProvider
		if stdErrors.As(err, &provider) {
			st := provider.GRPCStatus()
			if st != nil {
				return ParsedError{
					Result: ResultFailure,
					Code:   int(st.Code()),
					Msg:    st.Message(),
					Err:    err,
				}, true
			}
		}

		st, ok := grpcstatus.FromError(err)
		if !ok {
			return ParsedError{}, false
		}

		return ParsedError{
			Result: ResultFailure,
			Code:   int(st.Code()),
			Msg:    st.Message(),
			Err:    err,
		}, true
	})
}

