package errorx

// CodeError 表示 go-core 标准业务错误结构。
// 它既实现了标准 error 接口，也实现了 responsex 识别的
// ResponseCode/ResponseMsg 这两个“鸭子类型”方法，
// 因此 responsex.Builder 在默认配置下就能直接解析它。
type CodeError struct {
	Result ResultCode `json:"result"`
	Code   int        `json:"code"`
	Msg    string     `json:"msg"`
}

// CodeErrorResponse 是遗留的错误响应载荷结构。
// Deprecated: 新代码请使用 responsex.Builder.BuildError 构建统一响应。
type CodeErrorResponse struct {
	Result ResultCode `json:"result"`
	Code   int        `json:"code"`
	Msg    string     `json:"msg"`
}

func (e *CodeError) normalizedCode() int {
	if e == nil {
		return Unknown.Value()
	}

	if e.Result == ResultFailure && e.Code == OK.Value() {
		return Unknown.Value()
	}

	return e.Code
}

// NewCodeError 根据标准错误码构造失败错误。
func NewCodeError(code Code) error {
	return &CodeError{Result: ResultFailure, Code: code.Value(), Msg: code.Msg()}
}

// NewMsgError 根据自定义文案构造失败错误。
// 该场景会默认使用 Unknown 作为错误码。
func NewMsgError(msg string) error {
	return &CodeError{Result: ResultFailure, Code: Unknown.Value(), Msg: msg}
}

// NewCodeMsgError 根据原始数值和文案构造失败错误。
// 当业务方需要兼容历史错误码、或错误码尚未注册时可使用该方法。
func NewCodeMsgError(code int, msg string) error {
	return &CodeError{Result: ResultFailure, Code: code, Msg: msg}
}

// NewDefaultMsgError 使用 Unknown 错误码构造失败错误。
func NewDefaultMsgError(msg string) error {
	return NewCodeMsgError(Unknown.Value(), msg)
}


// NewDefaultError 构造一个 Unknown 错误码的默认失败错误。
func NewDefaultError() error {
	return NewCodeError(Unknown)
}

// Error 返回错误对象的文案。
// 若未显式设置 Msg，则会根据错误码回退到已注册的默认文案。
func (e *CodeError) Error() string {
	if e == nil {
		return Unknown.Msg()
	}
	if e.Msg == "" {
		return Code(e.normalizedCode()).Msg()
	}
	return e.Msg
}

// ResponseCode 让 CodeError 能被 responsex.DecodeResponseCoder 直接识别。
// responsex 不需要 import errorx，只需要看到 err 上有这个方法即可。
func (e *CodeError) ResponseCode() int {
	return e.normalizedCode()
}

// ResponseMsg 让 CodeError 能被 responsex.DecodeResponseCoder 直接识别。
func (e *CodeError) ResponseMsg() string {
	return e.Error()
}

// Data 返回遗留的错误响应结构。
// Deprecated: 新代码请直接使用 responsex.Builder.BuildError 生成统一响应。
func (e *CodeError) Data() *CodeErrorResponse {
	return &CodeErrorResponse{
		Result: ResultFailure,
		Code:   e.normalizedCode(),
		Msg:    e.Error(),
	}
}
