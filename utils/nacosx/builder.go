package nacosx

import (
	"fmt"
	"path/filepath"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type config struct {
	// 服务器配置
	IPAddr string // 服务地址
	Port   uint64 // 服务端口
	Path   string // 上下文路径

	// 客户端配置
	NamespaceID  string // 命名空间ID
	Username     string // 用户名
	Password     string // 密码
	TimeoutMs    uint64 // 超时时间(ms)
	LogDir       string // 日志目录
	CacheDir     string // 缓存目录
	LogLevel     string // 日志级别
	NotLoadCache bool   // 启动时是否不加载缓存
	DataID       string // 数据id

	// 服务注册配置
	ServiceName    string // 服务名称
	ServiceIP      string // 服务IP
	ServicePort    uint64 // 服务端口
	ServiceGroup   string // 服务分组
	ServiceCluster string // 服务集群
}

// Builder 配置构建器
type Builder struct {
	config *config
}

// NewBuilder 创建构建器并设置默认值
func NewBuilder() *Builder {
	return &Builder{
		config: &config{
			IPAddr:         "localhost",
			Port:           8848,
			Path:           "/nacos",
			LogDir:         filepath.Join("./log", "nacos"),
			CacheDir:       filepath.Join("./cache", "nacos"),
			LogLevel:       "error",
			NotLoadCache:   true,
			TimeoutMs:      5000,
			NamespaceID:    "",
			ServiceGroup:   "DEFAULT_GROUP",
			ServiceCluster: "DEFAULT",
		},
	}
}

// WithServerAddr 设置Nacos服务器地址和端口
func (b *Builder) WithServerAddr(ip string, port uint64) *Builder {
	b.config.IPAddr = ip
	b.config.Port = port
	return b
}

// WithContextPath 设置上下文路径
func (b *Builder) WithContextPath(path string) *Builder {
	b.config.Path = path
	return b
}

// WithNamespace 设置命名空间
func (b *Builder) WithNamespace(namespaceID string) *Builder {
	b.config.NamespaceID = namespaceID
	return b
}

// WithDataID
func (b *Builder) WithDataID(dataID string) *Builder {
	b.config.DataID = dataID
	return b
}

// WithAuth 设置认证信息
func (b *Builder) WithAuth(username, password string) *Builder {
	b.config.Username = username
	b.config.Password = password
	return b
}

// WithTimeout 设置超时时间(ms)
func (b *Builder) WithTimeout(timeoutMs uint64) *Builder {
	b.config.TimeoutMs = timeoutMs
	return b
}

// WithLogConfig 设置日志配置
func (b *Builder) WithLogConfig(logDir, logLevel string) *Builder {
	b.config.LogDir = logDir
	b.config.LogLevel = logLevel
	return b
}

// WithCacheConfig 设置缓存配置
func (b *Builder) WithCacheConfig(cacheDir string, notLoadCache bool) *Builder {
	b.config.CacheDir = cacheDir
	b.config.NotLoadCache = notLoadCache
	return b
}

// WithServiceInfo 设置服务注册信息
func (b *Builder) WithServiceInfo(name, ip string, port uint64) *Builder {
	b.config.ServiceName = name
	b.config.ServiceIP = ip
	b.config.ServicePort = port
	return b
}

// WithServiceGroup 设置服务分组
func (b *Builder) WithServiceGroup(group string) *Builder {
	b.config.ServiceGroup = group
	return b
}

// WithServiceCluster 设置服务集群
func (b *Builder) WithServiceCluster(cluster string) *Builder {
	b.config.ServiceCluster = cluster
	return b
}

// Execute 生成Nacosx实例
func (b *Builder) Execute() *Nacosx {
	nacosx := &Nacosx{config: b.config}

	// 创建服务器配置
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(
			b.config.IPAddr,
			b.config.Port,
			constant.WithContextPath(b.config.Path),
		),
	}

	// 创建客户端配置
	clientConfig := &constant.ClientConfig{
		NamespaceId:         b.config.NamespaceID,
		TimeoutMs:           b.config.TimeoutMs,
		NotLoadCacheAtStart: b.config.NotLoadCache,
		LogDir:              b.config.LogDir,
		CacheDir:            b.config.CacheDir,
		LogLevel:            b.config.LogLevel,
		Username:            b.config.Username,
		Password:            b.config.Password,
	}

	// 初始化命名客户端
	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create naming client: %w", err))
	}
	nacosx.namingClient = namingClient

	// 初始化配置客户端
	configClient, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create config client: %w", err))
	}
	nacosx.configClient = configClient

	// 如果提供了服务信息，则注册服务
	if b.config.ServiceName != "" && b.config.ServiceIP != "" && b.config.ServicePort > 0 {
		if err := nacosx.RegisterService(); err != nil {
			panic(fmt.Errorf("failed to register service: %w", err))
		}
	}

	return nacosx
}
