# opentelemetry 接入指南

## 前置步骤

```bash
# 在你的服务项目中引入 go-core
go get github.com/betacats/go-core@latest
```

---

## REST 服务接入（如 api-xd）

### 1. 修改 `cmd/cobra.go`

```go
func main() {
    // ... 已有 GetConfig 代码 ...

    if c, err = nacosx.GetConfig(); err != nil {
        panic(err)
    }

    // 关闭 go-zero 内置 tracing，改用 opentelemetry 方案
    c.Telemetry.Disabled = true

    // 初始化自定义 tracing
    cfg := opentelemetry.NewConfig(
        c.Telemetry.Name,
        c.Telemetry.Endpoint,                             // "http://arms-dc-xxx.aliyuncs.com:443"
        c.Telemetry.OtlpHttpPath,                         // "/v1/traces"
    )
    cfg.Headers = map[string]string{
        "ARMS-USERNAME": c.Telemetry.Endpoint,
        "ARMS-PASSWORD": c.Telemetry.Endpoint,
    }
    cfg.NormalSampler = 0.1                               // 正常 span 10% 采样

    shutdown, err := opentelemetry.InitTracing(cfg)
    if err != nil {
        panic(err)
    }
    proc.AddShutdownListener(func() { _ = shutdown(context.Background()) })

    // ... 继续启动服务 ...
}
```

### 2. 修改 `cmd/middleware.go` — 注册中间件

> 顺序决定：所有中间件执行完 → `BodyRecorder` 截获完整响应 → 判断业务错误

```go
server.Use(MiddlewareWithSentry())        // panic 恢复 → 写错误 body
server.Use(middlewareWithRedisToken(ctx))  // token 校验
server.Use(func(next http.HandlerFunc) http.HandlerFunc {
    return opentelemetry.Middleware(next).ServeHTTP  // 拦截 body 标记 Error ← 必须最后
})
```

### 3. handler 中使用统—响应

handler 无需感知 tracing，只需正常返回 error：

```go
// internal/logic/create_order.go
func (l *CreateOrderLogic) CreateOrder(req *types.Req) (*types.Resp, error) {
    if req.Sku == "" {
        return nil, errorx.NewCodeError(errorx.InvalidArgument)
    }
    return &types.Resp{OrderNo: "SO123"}, nil
}
```

```go
// internal/handler/create_order_handler.go
func CreateOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.Req
        if err := httpx.Parse(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }
        l := logic.NewCreateOrderLogic(r.Context(), svcCtx)
        resp, err := l.CreateOrder(&req)
        if err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }
        httpx.OkJsonCtx(r.Context(), w, resp)
    }
}
```

`httpx.ErrorCtx` / `httpx.OkJsonCtx` 输出响应 → `opentelemetry.Middleware` 截获 body → 检测 `result:false` → 标记 span Error → 整条 trace 全量上报 ARMS。

---

## gRPC 服务接入（如 rpc-user）

### 1. 修改 `cmd/cobra.go`（同上）

```go
c.Telemetry.Disabled = true
c.Telemetry.Sampler = 1.0

cfg := opentelemetry.NewConfig(
    c.Telemetry.Name,
    c.Telemetry.Endpoint,
    c.Telemetry.OtlpHttpPath,
)

shutdown, err := opentelemetry.InitTracing(cfg)
if err != nil {
    panic(err)
}
proc.AddShutdownListener(func() { _ = shutdown(context.Background()) })
```

### 2. 不需要额外中间件

gRPC 的 `UnaryTracingInterceptor` 会自动：
- handler 返回 error → 框架自动 `SetStatus(codes.Error)`
- 调下游 RPC 报错 → 客户端拦截器自动标 ClientSpan Error

### 3. handler 返回 error 即可

```go
func (l *CreateOrderLogic) CreateOrder(in *pb.CreateOrderReq) (*pb.CreateOrderResp, error) {
    if in.Sku == "" {
        return nil, errorx.NewCodeError(errorx.InvalidArgument)
    }
    return &pb.CreateOrderResp{OrderNo: "SO123"}, nil
}
```

---

## Config 字段速查

| 字段 | 默认值 | 说明 |
|---|---|---|
| `ServiceName` | **必填** | 服务名，写入 span resource |
| `Endpoint` | **必填** | OTLP collector 地址（ARMS） |
| `URLPath` | **必填** | OTLP HTTP 路径 |
| `NormalSampler` | 0.1 | 正常 span 采样率 [0,1] |
| `BatchTimeout` | 5 | 批量导出最大等待秒数 |
| `MaxExportBatchSize`` | 512 | 单次批量最大 span 数 |
| `Insecure` | false | 跳过 TLS |
| `Headers` | nil | OTLP 请求附加 Header（ARMS 认证） |
| `LRUMaxSize` | 10000 | 错误 traceID 缓存上限 |
| `ErrorTTLSeconds` | 30 | 错误 traceID 保留秒数 |

---

## 验证是否接入成功

1. 请求一次正常接口 → 检查 ARMS 中该 trace 的采样率为 10%
2. 请求一次返回 `result:false` 的接口 → 检查 ARMS 中该 trace 100% 保留
3. 调下游 RPC 并让下游报错 → 检查上下游 span 在 ARMS 中拼接为完整调用链
