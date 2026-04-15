package restyx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Option 定义客户端配置选项
type Option func(*Client)

// Client 封装 resty.Client，兼容原有方法并集成 OTEL
type Client struct {
	*resty.Client
	tracer trace.Tracer // OTEL 追踪器
}

// NewClient 创建带 OTEL 追踪的 resty 客户端
// 支持通过 Option 自定义配置
func NewClient(opts ...Option) *Client {
	// 初始化基础 resty 客户端
	restyClient := resty.New()

	// 初始化默认 OTEL 配置
	client := &Client{
		Client: restyClient,
		tracer: otel.Tracer("resty-otel"), // 追踪器名称
	}

	// 应用用户自定义选项
	for _, opt := range opts {
		opt(client)
	}

	// 配置默认 OTEL 传输层（带请求体追踪）
	client.setDefaultOTELTransport()

	return client
}

// 配置默认 OTEL 传输层，自动追踪请求
func (c *Client) setDefaultOTELTransport() {
	transport := otelhttp.NewTransport(http.DefaultTransport,
		otelhttp.WithPropagators(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)),
	)
	c.SetTransport(transport)
}

// Execute 执行 HTTP 请求并自动记录 OTEL 追踪
// 封装 resty 的 R() 方法，自动处理 span 生命周期
func (c *Client) Execute(req *resty.Request, method, url string) (*resty.Response, error) {
	// 从请求中提取信息用于追踪
	reqBody, _ := json.Marshal(c.parseData(req))
	ctx := req.Context()

	// 创建 span
	ctx, span := c.tracer.Start(ctx, fmt.Sprintf("%s %s", method, url),
		trace.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.url", url),
			attribute.String("req.body", string(reqBody)),
		),
	)
	defer span.End()

	// 将 ctx 注入请求（用于传递追踪上下文）
	req.SetContext(ctx)

	// 执行请求
	var resp *resty.Response
	var err error
	switch method {
	case http.MethodGet:
		resp, err = req.Get(url)
	case http.MethodPost:
		resp, err = req.Post(url)
	case http.MethodPut:
		resp, err = req.Put(url)
	case http.MethodDelete:
		resp, err = req.Delete(url)
	case http.MethodOptions:
		resp, err = req.Options(url)
	case http.MethodPatch:
		resp, err = req.Patch(url)
	case http.MethodHead:
		resp, err = req.Head(url)
	default:
		err = fmt.Errorf("unsupported http method: %s", method)
	}

	// 获取当前 span 的上下文
	spanCtx := span.SpanContext()

	// 获取 TraceID
	traceID := spanCtx.TraceID().String()

	// 记录响应信息到 span
	if resp != nil {
		span.SetAttributes(
			attribute.Int("http.status_code", resp.StatusCode()),
			attribute.String("resp.body", string(resp.Body())),
		)
		if resp.RawResponse != nil {
			resp.RawResponse.Header.Set("X-Trace-ID", traceID)
		}
	}

	// 记录错误
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return resp, err
}

// parseData 解析请求数据
func (c *Client) parseData(req *resty.Request) (requestData map[string]interface{}) {
	requestData = map[string]interface{}{
		"header": req.Header,
		"query":  req.QueryParam,
		"form":   req.FormData,
		"raw":    req.RawPathParams,
		"path":   req.PathParams,
		"body":   req.Body,
	}

	return requestData
}

// SetTransportConfig 配置HTTP transport
func (c *Client) SetTransportConfig(config *http.Transport) *Client {
	otelTr := otelhttp.NewTransport(config,
		otelhttp.WithPropagators(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)),
	)
	c.SetTransport(otelTr)
	return c
}

// Option 选项：设置重试配置
func WithRetry(count int, waitTime, maxWaitTime time.Duration) Option {
	return func(c *Client) {
		if count < 0 {
			return
		}
		c.SetRetryCount(count)
		c.SetRetryWaitTime(waitTime)
		c.SetRetryMaxWaitTime(maxWaitTime)
	}
}

// Option 选项：自定义 OTEL 追踪器名称
func WithTracerName(name string) Option {
	return func(c *Client) {
		c.tracer = otel.Tracer(name)
	}
}

// Option 选项：设置请求超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout > 0 {
			c.SetTimeout(timeout)
		}
	}
}

// Option 选项：配置请求头
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) {
		if headers != nil {
			for k, v := range headers {
				c.SetHeader(k, v)
			}
		}
	}
}

// Option 选项：配置基础地址
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL != "" {
			c.SetBaseURL(baseURL)
		}
	}
}
