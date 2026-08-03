# asyncengine

进程内异步任务引擎：`SubmitFunc(闭包)` + Redis 结果存储 + worker 池。

## 接入

```go
runner := asyncengine.MustNew(redisClient, asyncengine.Config{
    Workers:     4,
    TaskTimeout: 10 * time.Minute,
    OnComplete:  onComplete, // 可选：SSE / 通知
})
defer runner.Shutdown()

taskID, err := runner.SubmitFunc(ctx, func(ctx context.Context) (asyncengine.Result, error) {
    // 业务逻辑，例如调 RPC
    return asyncengine.Result{Data: map[string]any{
        "total": 100,
        "count": 3,
        "url":   "https://...",
    }}, nil
}, asyncengine.NotifyTarget{
    HospitalNo: "...",
    ShopNo:     "...",
    EmployeeNo: "...", // 必填：通知人
})

record, err := runner.GetRecord(ctx, taskID)
// record.State: PENDING / RUNNING / SUCCEEDED / FAILED
// record.Result.Data: 成功时的业务结果（结构由提交方约定）
// record.Error: 失败原因
```

## 能力

- 状态机：`PENDING`（排队）→ `RUNNING`（执行中）→ `SUCCEEDED` / `FAILED`
- 启动对账：重启后仍为 `PENDING` / `RUNNING` 的任务标为失败
- channel 满时返回 `ErrQueueFull`（不阻塞 HTTP）
- `OnComplete` 终态回调（重启对账导致的失败不触发）
- 业务结果统一放在 `Result.Data`，查询侧原样透传，由调用方约定结构

## Redis 键值

前缀由 `Config.KeyPrefix` 决定，默认 `asyncengine`（api-pet 使用 `api-pet`）。

| Key | 类型 | TTL | 说明 |
|---|---|---|---|
| `{prefix}:result:{taskId}` | String（JSON） | 24h | 任务完整记录 |
| `{prefix}:pending` | Set | 无 | 当前未终态（`PENDING` / `RUNNING`）的 taskId 集合，供启动对账 |

`{prefix}:result:{taskId}` value 示例：

```json
{
  "state": "SUCCEEDED",
  "result": {
    "data": {
      "total": 100,
      "count": 3,
      "url": "https://..."
    }
  },
  "notify": {
    "hospitalNo": "...",
    "shopNo": "...",
    "employeeNo": "..."
  }
}
```

- `state`：`PENDING`（排队中）/ `RUNNING`（执行中）/ `SUCCEEDED` / `FAILED`
- `result.data`：成功时写入，业务结果结构由提交方约定；失败时可省略
- `error`：失败原因；成功时可省略
- `notify`：提交时绑定的归属，用于查询鉴权与 OnComplete 路由

生命周期：提交时写 result（`PENDING`）+ `SADD pending` → worker 取走后覆盖写 `RUNNING` → 终态覆盖 result（`SUCCEEDED`/`FAILED`）并 `SREM pending` → result 过期后自动消失。
