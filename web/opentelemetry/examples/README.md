# opentelemetry 接入指南

## 前置步骤

```bash
# 在你的服务项目中引入 go-core
go get github.com/betacats/go-core@latest
```

---

## REST / gRPC 服务接入

### 1. 修改 `cmd/cobra.go`

```go
func main() {
    // ... 已有 GetConfig 代码 ...

    if c, err = nacosx.GetConfig(); err != nil {
        panic(err)
    }

    // 关闭 go-zero 内置 tracing，改用 opentelemetry 方案
    c.Telemetry.Disabled = true

    // cfg.NormalSampler 零值时 getter 返回 0.1（默认 10%）
    shutdown, err := opentelemetry.InitTracing(opentelemetry.Config{
        ServiceName: c.Telemetry.Name,
        Endpoint:    c.Telemetry.Endpoint,
        URLPath:     c.Telemetry.OtlpHttpPath,
    })
    if err != nil {
        panic(err)
    }
    proc.AddShutdownListener(func() { _ = shutdown(context.Background()) })

    // ... 继续启动服务 ...
}
```

### 2. 在 handler/logic 中标记 span

```go
// internal/handler/create_order_handler.go
func (l *CreateOrderLogic) CreateOrder(req *types.Req) (*types.Resp, error) {
    resp, err := l.svcCtx.OrderModel.Insert(l.ctx, req)
    if err != nil {
        return nil, err
    }
    return &types.Resp{OrderNo: resp.OrderNo}, nil
}
```

gRPC handler 返回 error 时，go-zero 的 `UnaryTracingInterceptor` 自动标记 span Error。

如果需要手动标记：

```go
// 主动标记为错误（触发全量保留）
opentelemetry.MarkSpanError(ctx, responsex.Response{
    Result: responsex.ResultFailure,
    Code:   7,
    Msg:    "business error",
})

// 主动标记为成功
opentelemetry.MarkSpanOk(ctx, responsex.Response{
    Result: responsex.ResultSuccess,
    Code:   0,
    Msg:    "success",
})
```

---

## Config 字段速查

### 必填字段

| 字段 | 说明 |
|---|---|
| `ServiceName` | 服务名，写入 span resource |
| `Endpoint` | OTLP collector 地址（ARMS） |
| `URLPath` | OTLP HTTP 路径 |

### 可选字段

| 字段 | Getter 默认值 | 说明 |
|---|---|---|
| `NormalSampler` | `GetNormalSampler()` → 0.1 | 正常 span 采样率 [0,1] |
| `LRUMaxSize` | `GetLRUMaxSize()` → 10000 | 错误 traceID 缓存上限 |
| `ErrorTTLSeconds` | `GetErrorTTLSeconds()` → 30 | 错误 traceID 保留秒数 |
| `BatchTimeout` | `GetBatchTimeout()` → 5 | 批量导出最大等待秒数 |
| `MaxExportBatchSize` | `GetMaxExportBatchSize()` → 512 | 单次批量最大 span 数 |
| `Batcher` | `GetBatcher()` → "batch" | 保留供后续扩展 |
| `Insecure` | `GetInsecure()` → false | 跳过 TLS |
| `Headers` | `GetHeaders()` → nil | OTLP 请求附加 Header |

---

## API 一览

| 函数 | 说明 |
|---|---|
| `InitTracing(cfg Config)` | 初始化全局 TracerProvider |
| `StopTracing()` | 优雅关闭 |
| `MarkSpanError(ctx, resp)` | 将当前 span 标记为 Error（全量保留） |
| `MarkSpanOk(ctx, resp)` | 将当前 span 标记为 Ok |

---

## 验证是否接入成功

1. 请求一次正常接口 → 检查 ARMS 中该 trace 的采样率为 10%
2. 请求一次返回 `result:false` 的接口 → 检查 ARMS 中该 trace 100% 保留
3. 调下游 RPC 并让下游报错 → 检查上下游 span 在 ARMS 中拼接为完整调用链
