# responsex examples

这个目录放 **偏集成场景** 的示例，例如：
- HTTP handler 接入
- 中间件写入 `RequestMeta`
- 默认 `SentryReporter` 上报链路

建议约定：
- `web/responsex/` 根目录只保留 **短小、直接对应公开 API 的 GoDoc Example**
- `web/responsex/examples/` 放 **更完整的端到端示例**，避免包根目录示例文件过多

当前示例：
- `http_integration_test.go`：演示在 HTTP handler 中使用 `responsex.Builder`

