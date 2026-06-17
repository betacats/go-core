package errorx

import "testing"

func TestCodeErrorResponseAccessors(t *testing.T) {
	t.Parallel()

	err := &CodeError{Result: ResultFailure, Code: PermissionDenied.Value()}
	if err.ResponseCode() != PermissionDenied.Value() {
		t.Fatalf("expected response code %d, got %d", PermissionDenied.Value(), err.ResponseCode())
	}
	if err.ResponseMsg() != PermissionDenied.Msg() {
		t.Fatalf("expected response msg %q, got %q", PermissionDenied.Msg(), err.ResponseMsg())
	}
}

func TestCodeErrorFailureWithOKFallsBackToUnknown(t *testing.T) {
	t.Parallel()

	err := &CodeError{Result: ResultFailure, Code: OK.Value(), Msg: "bad request"}
	if err.ResponseCode() != Unknown.Value() {
		t.Fatalf("expected response code %d, got %d", Unknown.Value(), err.ResponseCode())
	}
	if err.ResponseMsg() != "bad request" {
		t.Fatalf("expected original msg to be preserved, got %q", err.ResponseMsg())
	}
}

func TestRegisterCode(t *testing.T) {
	const customCode Code = 990001
	const customMsg = "自定义业务错误"

	if got := FindMsg(customCode); got != "" {
		t.Fatalf("expected code %d to be unregistered before test, got %q", customCode, got)
	}

	if err := RegisterCode(customCode, customMsg); err != nil {
		t.Fatalf("register code: %v", err)
	}
	if got := customCode.Msg(); got != customMsg {
		t.Fatalf("expected registered msg %q, got %q", customMsg, got)
	}
	if got := FindCode(customMsg); got != customCode {
		t.Fatalf("expected FindCode to return %d, got %d", customCode, got)
	}
	if err := RegisterCode(customCode, "重复注册"); err == nil {
		t.Fatal("expected duplicate registration to fail")
	}
}

func TestMustRegisterCode(t *testing.T) {
	const customCode Code = 990004
	const customMsg = "必须注册成功"

	MustRegisterCode(customCode, customMsg)

	if got := customCode.Msg(); got != customMsg {
		t.Fatalf("expected registered msg %q, got %q", customMsg, got)
	}
}

func TestNewMsgError(t *testing.T) {
	t.Parallel()

	err, ok := NewMsgError("plain error").(*CodeError)
	if !ok {
		t.Fatal("expected *CodeError")
	}
	if err.Code != Unknown.Value() {
		t.Fatalf("expected unknown code %d, got %d", Unknown.Value(), err.Code)
	}
	if err.Msg != "plain error" {
		t.Fatalf("expected msg %q, got %q", "plain error", err.Msg)
	}
}

func TestNewDefaultError(t *testing.T) {
	t.Parallel()

	err, ok := NewDefaultError().(*CodeError)
	if !ok {
		t.Fatal("expected *CodeError")
	}
	if err.Code != Unknown.Value() {
		t.Fatalf("expected unknown code %d, got %d", Unknown.Value(), err.Code)
	}
	if err.Msg != Unknown.Msg() {
		t.Fatalf("expected msg %q, got %q", Unknown.Msg(), err.Msg)
	}
}

