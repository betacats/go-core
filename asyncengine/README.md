# asyncengine

进程内异步任务引擎：`SubmitFunc(闭包)` + Redis 结果存储 + worker 池。

状态机：`PENDING`（排队）→ `RUNNING`（执行中）→ `SUCCEEDED` / `FAILED`

---

## 以 api-pet 为例：如何接入

下面按真实服务接入步骤说明（go-zero + cobra + Redis + SSE 通知）。

### 1. 依赖

```bash
go get github.com/betacats/go-core@最新版本
```

```go
import "github.com/betacats/go-core/asyncengine"
```

### 2. ServiceContext 挂上 Runner

`internal/svc/service_context.go`：

```go
type ServiceContext struct {
    // ...
    DefaultRdsEngine *redis.Client
    AsyncRunner      *asyncengine.Runner // 新增
}
```

### 3. 新建 `internal/async/init.go`

负责创建 Runner、优雅关闭，以及任务完成时的通知（api-pet 走现有 SSE）：

```go
package async

import (
    "context"
    "encoding/json"
    "time"

    "github.com/betacats/go-core/asyncengine"
    "github.com/zeromicro/go-zero/core/logx"

    // 按你们项目实际路径调整
    "api-pet/internal/svc"
    notifylogic "api-pet/internal/logic/v1/notify"
    "api-pet/internal/types"
)

type logxAdapter struct{}

func (logxAdapter) Infof(format string, args ...any)  { logx.Infof(format, args...) }
func (logxAdapter) Errorf(format string, args ...any) { logx.Errorf(format, args...) }

// MustInit 初始化异步任务 Runner。
func MustInit(svcCtx *svc.ServiceContext) {
    svcCtx.AsyncRunner = asyncengine.MustNew(svcCtx.DefaultRdsEngine, asyncengine.Config{
        Workers:     4,
        TaskTimeout: 10 * time.Minute,
        KeyPrefix:   "api-pet", // 多服务共用 Redis 时务必区分前缀
        Logger:      logxAdapter{},
        OnComplete:  makeOnComplete(svcCtx), // 可选：完成后推 SSE
    })
}

func Shutdown(svcCtx *svc.ServiceContext) {
    if svcCtx.AsyncRunner != nil {
        svcCtx.AsyncRunner.Shutdown()
    }
}

func makeOnComplete(svcCtx *svc.ServiceContext) func(asyncengine.Completion) {
    return func(c asyncengine.Completion) {
        if c.Notify.EmployeeNo == "" {
            return
        }
        status := asyncengine.StateSucceeded
        errMsg := ""
        var data any
        title := "异步任务已完成"
        if c.Err != nil {
            status = asyncengine.StateFailed
            errMsg = c.Err.Error()
            title = "异步任务失败"
        } else {
            data = c.Result.Data
        }
        ext, _ := json.Marshal(map[string]any{
            "kind":   "async.task",
            "taskId": c.TaskID,
            "status": status,
            "data":   data,
            "error":  errMsg,
        })
        _ = notifylogic.NewSendLogic(context.Background(), svcCtx).Send(&types.NotifySendReq{
            Title:    title,
            Content:  title,
            PushTime: time.Now().Unix(),
            ExtBody:  string(ext),
            NotifyEmployee: []types.NotifyEmployee{{
                EmployeeNo: c.Notify.EmployeeNo,
                HospitalNo: c.Notify.HospitalNo,
                ShopNo:     c.Notify.ShopNo,
            }},
        })
    }
}
```

### 4. 在 `cmd/cobra.go` 启动 / 关闭

在创建 `ServiceContext` 之后初始化，进程退出前关闭：

```go
svcCtx := svc.NewServiceContext(svcConf)
async.MustInit(svcCtx)
defer async.Shutdown(svcCtx)

initServer(&svcConf, svcCtx)
```

> 注意：`Shutdown` 应在 HTTP server 停止之后执行（`defer` 按 LIFO，保证先停服、再停 Runner），避免关闭过程中还有请求往队列里提交。

### 5. 业务里提交任务（例如商品导入）

接口立即返回 `taskId`，耗时逻辑放进闭包：

```go
fn := func(ctx context.Context) (asyncengine.Result, error) {
    reply, err := l.svcCtx.RPCGoods.ImportNomalGoods(ctx, importReq)
    if err != nil {
        return asyncengine.Result{}, err
    }
    // 业务结果放进 Data，结构由调用方约定，查询接口原样透传
    return asyncengine.Result{Data: map[string]any{
        "total": reply.Total,
        "count": reply.Count,
        "url":   reply.Url,
    }}, nil
}

taskID, err := l.svcCtx.AsyncRunner.SubmitFunc(l.ctx, fn, asyncengine.NotifyTarget{
    HospitalNo: l.svcCtx.GetHospitalNo(l.ctx),
    ShopNo:     l.svcCtx.GetShopNo(l.ctx),
    EmployeeNo: l.svcCtx.GetEmployeeNo(l.ctx), // 必填：通知人 / 查询鉴权
})
if errors.Is(err, asyncengine.ErrQueueFull) {
    return nil, /* 资源繁忙 */ nil
}
// 返回 { "taskId": taskID }
```

### 6. 提供查询接口

前端用 `taskId` 轮询状态。api-pet 示例：`GET /v1/async/task?taskId=xxx`

逻辑要点：

1. `runner.GetRecord(ctx, taskId)` 取完整记录
2. 用 `record.Notify` 做归属校验（医院 / 门店 / 员工）
3. 返回：

```json
{
  "taskId": "...",
  "status": "SUCCEEDED",
  "error": "",
  "data": {
    "total": 100,
    "count": 3,
    "url": "https://..."
  }
}
```

`data` 原样透传 `record.Result.Data`，不要在查询层写死业务结构。前端按自己发起的任务类型解析即可。

### 7. 接入清单（复制用）

| 步骤 | 文件 | 做什么 |
|---|---|---|
| 1 | `go.mod` | 依赖 `github.com/betacats/go-core` |
| 2 | `internal/svc/service_context.go` | 增加 `AsyncRunner *asyncengine.Runner` |
| 3 | `internal/async/init.go` | `MustInit` / `Shutdown` / 可选 `OnComplete` |
| 4 | `cmd/cobra.go` | `async.MustInit` + `defer async.Shutdown` |
| 5 | 业务 logic | `SubmitFunc` 提交闭包，返回 `taskId` |
| 6 | 查询 API | `GetRecord` + 归属校验 + 透传 `data` |

---

## API 速览

```go
runner := asyncengine.MustNew(redisClient, asyncengine.Config{
    Workers:     4,
    TaskTimeout: 10 * time.Minute,
    KeyPrefix:   "api-pet",
    OnComplete:  onComplete, // 可选
    Logger:      logger,     // 可选
})
defer runner.Shutdown()

taskID, err := runner.SubmitFunc(ctx, fn, asyncengine.NotifyTarget{
    EmployeeNo: "...", // 必填
})

record, err := runner.GetRecord(ctx, taskID)
// record.State / record.Result.Data / record.Error / record.Notify
```

---

## 能力

- 状态机：`PENDING` → `RUNNING` → `SUCCEEDED` / `FAILED`
- 启动对账：重启后仍为 `PENDING` / `RUNNING` 的任务标为失败（不触发 `OnComplete`）
- channel 满时返回 `ErrQueueFull`（不阻塞 HTTP）
- `OnComplete` 终态回调
- 业务结果统一放在 `Result.Data`，查询侧原样透传

---

## Redis 键值

前缀由 `Config.KeyPrefix` 决定，默认 `asyncengine`。多服务共用 Redis 时务必设置不同前缀。

| Key | 类型 | TTL | 说明 |
|---|---|---|---|
| `{prefix}:result:{taskId}` | String（JSON） | 24h | 任务完整记录 |
| `{prefix}:pending` | Set | 无 | 未终态（`PENDING` / `RUNNING`）taskId，供启动对账 |

生命周期：提交写 `PENDING` + `SADD pending` → worker 取走写 `RUNNING` → 终态覆盖并 `SREM pending` → result 过期消失。

---

## 单测

同目录 `example_test.go`（miniredis）覆盖提交成功、失败回调、启动对账：

```bash
go test ./asyncengine/ -v
```
