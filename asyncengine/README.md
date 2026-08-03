# asyncengine

进程内异步任务引擎：`SubmitFunc(闭包)` + Redis 状态存储 + worker 池。

状态机：`PENDING`（排队）→ `RUNNING`（执行中）→ `SUCCEEDED` / `FAILED`

---

## 快速使用

```go
import (
    "context"
    "time"

    "github.com/betacats/go-core/asyncengine"
    "github.com/redis/go-redis/v9"
)

runner := asyncengine.MustNew(redisClient, asyncengine.Config{
    Namespace:   "my-service", // 必填，按服务区分 Redis key
    Workers:     4,
    TaskTimeout: 10 * time.Minute,
    OnComplete: func(c asyncengine.Completion) {
        // 可选：任务终态回调（成功或失败）
        // c.TaskID / c.Result / c.Err / c.Meta
    },
})
defer runner.Shutdown()

taskID, err := runner.SubmitFunc(ctx, func(ctx context.Context) (asyncengine.Result, error) {
    // 耗时逻辑放在闭包里
    return asyncengine.Result{Data: map[string]any{
        "url": "https://example.com/file.xlsx",
    }}, nil
}, map[string]string{
    "ownerId": "u001", // 可选元数据，引擎原样存储与回传
})
if err != nil {
    // 处理 ErrQueueFull / Redis 写入失败等
}

record, err := runner.GetRecord(ctx, taskID)
// record.State: PENDING / RUNNING / SUCCEEDED / FAILED
// record.Result.Data / record.Error / record.Meta
```

---

## 典型接入模式

适用于 HTTP 服务：提交接口立即返回 `taskId`，客户端轮询查询接口。

### 1. 进程启动时创建 Runner

```go
runner := asyncengine.MustNew(rdb, asyncengine.Config{
    Namespace:   "api-pet", // 必填，同一服务多实例用相同值
    Workers:     4,
    TaskTimeout: 10 * time.Minute,
    OnComplete:  onTaskComplete, // 可选，应快速返回
})
defer runner.Shutdown()
```

> `Shutdown` 应在 HTTP server 停止之后执行，避免关闭过程中仍有新任务提交。

### 2. 提交任务

```go
taskID, err := runner.SubmitFunc(ctx, func(ctx context.Context) (asyncengine.Result, error) {
    result, err := doHeavyWork(ctx)
    if err != nil {
        return asyncengine.Result{}, err
    }
    return asyncengine.Result{Data: result}, nil
}, map[string]string{
    "ownerId": ownerID, // 查询鉴权、通知路由等由调用方自行约定
})
```

队列满时返回 `ErrQueueFull`，调用方可映射为 429/503 等。

### 3. 查询任务

```go
record, err := runner.GetRecord(ctx, taskID)
if record == nil {
    // 任务不存在
}
// 业务侧按 record.Meta 做归属校验
// 按 record.State / record.Error / record.Result.Data 组装响应
```

查询响应示例（字段名由业务 API 自行定义）：

```json
{
  "taskId": "...",
  "status": "SUCCEEDED",
  "error": "",
  "data": { "url": "https://..." }
}
```

### 接入清单

| 步骤 | 说明 |
|---|---|
| 1 | 进程启动：`MustNew`，退出前 `Shutdown` |
| 2 | 提交：`SubmitFunc(ctx, fn, meta)`，立即返回 `taskId` |
| 3 | 查询：`GetRecord(ctx, taskId)`，按 `Meta` 鉴权、透传 `Data` |
| 4 | 可选：`OnComplete` 做终态通知（回调内勿阻塞） |

---

## 能力

- 状态机：`PENDING` → `RUNNING` → `SUCCEEDED` / `FAILED`
- 启动对账：重启后仍为 `PENDING` / `RUNNING` 的任务标为失败（不触发 `OnComplete`）
- 队列满时返回 `ErrQueueFull`（非阻塞提交）
- `OnComplete` 终态回调
- `Meta` / `Result.Data` 由调用方约定，引擎原样透传
- 成功结果写 Redis 失败时，降级标 `FAILED`（`error=保存任务结果失败`）

---

## 使用边界

- **单进程队列**：任务闭包仅在当前进程 worker 内执行，不能跨进程恢复
- **Namespace 必填**：不同服务使用不同 `Namespace`，避免共用 pending 集合导致启动对账互相误标
- **OnComplete 应轻量**：在 worker goroutine 内同步调用，耗时逻辑请自行异步化

---

## Redis 键值

键名格式：`asyncengine:{Namespace}:...`，`Namespace` 由 `Config.Namespace` 传入（如 `api-pet`）。

| Key | 类型 | TTL | 说明 |
|---|---|---|---|
| `asyncengine:{ns}:result:{taskId}` | String（JSON） | 24h | 任务完整记录 |
| `asyncengine:{ns}:pending` | Set | 无 | 未终态 taskId，供启动对账 |

记录示例：

```json
{
  "state": "SUCCEEDED",
  "result": { "data": { "url": "https://..." } },
  "meta": { "ownerId": "u001" }
}
```

生命周期：提交写 `PENDING` + `SADD pending` → worker 取走写 `RUNNING` → 终态覆盖并 `SREM pending` → result 过期消失。

---

