package responsex_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	"github.com/betacats/go-core/web/errorx"
	"github.com/betacats/go-core/web/responsex"
)

func TestBuildErrorWithGRPCStatus(t *testing.T) {
	t.Parallel()

	builder := newBusinessErrorBuilder()
	resp := builder.BuildError(context.Background(), grpcstatus.Error(grpccodes.NotFound, "rpc resource not found"))

	if resp.Result != responsex.ResultFailure {
		t.Fatalf("expected failure result, got %v", resp.Result)
	}
	if resp.Code != int(grpccodes.NotFound) {
		t.Fatalf("expected code %d, got %d", grpccodes.NotFound, resp.Code)
	}
	if resp.Msg != "rpc resource not found" {
		t.Fatalf("expected msg %q, got %q", "rpc resource not found", resp.Msg)
	}
}

func TestBuildErrorWithWrappedGRPCStatus(t *testing.T) {
	t.Parallel()

	builder := newBusinessErrorBuilder()
	err := fmt.Errorf("call downstream failed: %w", grpcstatus.Error(grpccodes.PermissionDenied, "rpc no permission"))
	resp := builder.BuildError(context.Background(), err)

	if resp.Code != int(grpccodes.PermissionDenied) {
		t.Fatalf("expected code %d, got %d", grpccodes.PermissionDenied, resp.Code)
	}
	if resp.Msg != "rpc no permission" {
		t.Fatalf("expected msg %q, got %q", "rpc no permission", resp.Msg)
	}
}

func TestBuildErrorPrefersErrorxCodeErrorOverGRPCStatus(t *testing.T) {
	t.Parallel()

	builder := newBusinessErrorBuilder()
	resp := builder.BuildError(context.Background(), errorx.NewCodeError(errorx.PermissionDenied))

	if resp.Code != errorx.PermissionDenied.Value() {
		t.Fatalf("expected errorx code %d, got %d", errorx.PermissionDenied.Value(), resp.Code)
	}
	if resp.Msg != errorx.PermissionDenied.Msg() {
		t.Fatalf("expected errorx msg %q, got %q", errorx.PermissionDenied.Msg(), resp.Msg)
	}
}

func ExampleBuilder_withGRPCStatusError() {
	builder := newBusinessErrorBuilder()
	resp := builder.BuildError(context.Background(), grpcstatus.Error(grpccodes.Unavailable, "rpc downstream unavailable"))

	body, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
	// Output:
	// {"result":false,"code":14,"msg":"rpc downstream unavailable","data":{}}
}
