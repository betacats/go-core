package redis

import (
	"context"
	"fmt"
	"time"

	rd "github.com/redis/go-redis/v9"
)

// CreateClient create a client with option
func CreateClient(ctx context.Context, opt *RedisOption) (*rd.Client, error) {
	opts := &rd.Options{
		Addr:            opt.Addr,
		Username:        opt.Username,
		Password:        opt.Password,
		DB:              opt.DB,
		MaxRetries:      opt.MaxRetries,
		PoolSize:        opt.PoolSize,
		MinIdleConns:    opt.MinIdleConns,
		DialTimeout:     time.Millisecond * time.Duration(opt.ConnectTimeout),
		ReadTimeout:     time.Millisecond * time.Duration(opt.ReadTimeout),
		WriteTimeout:    time.Millisecond * time.Duration(opt.WriteTimeout),
		ConnMaxIdleTime: time.Second * time.Duration(opt.IdleTimeout),
	}

	client := rd.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("failed to connect to goRds: %v", err))
	}

	return client, nil
}

type RedisOption struct {
	Addr           string
	Username       string
	Password       string
	DB             int
	MinIdleConns   int // 最小空闲连接： 推荐值: 5-20 （说明: 预热连接池，减少首次请求延迟）
	MaxRetries     int // 最大重试次数：1-2: 推荐值，避免雪崩
	ConnectTimeout int // 连接超时时间 推荐值: 1000-5000 毫秒 (1-5秒)
	ReadTimeout    int // 读超时时间 推荐值: 3000-10000 毫秒 (3-10秒)
	WriteTimeout   int // 写超时时间 推荐值: 3000-10000 毫秒 (3-10秒)
	PoolSize       int // 连接池大小 推荐值: 20-100 (CPU核心数 * 2 ~ * 4)
	IdleTimeout    int // 空闲超时时间 推荐值: 60-300 秒
}
