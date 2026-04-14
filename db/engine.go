package db

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// createEngine create a new engine with option
func createEngine(opt *DataBaseOption, config gorm.Config) (*gorm.DB, error) {

	db, err := gorm.Open(mysql.Open(opt.Dsn), &config)

	if err != nil {
		return nil, err
	}
	sqlDb, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDb.SetMaxOpenConns(opt.OpenConnections)
	sqlDb.SetMaxIdleConns(opt.IdleConnections)
	sqlDb.SetConnMaxLifetime(time.Duration(opt.Lifetime) * time.Second)

	err = sqlDb.Ping() // Ping the database to ensure the connection is valid
	if err != nil {
		return nil, err
	}

	return db, nil
}

type DataBaseOption struct {
	Dsn             string // 数据库连接字符串 格式: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	IdleConnections int    // 最大空闲连接数 推荐值: 10-50 (说明: 保持适量空闲连接，减少连接创建开销)
	OpenConnections int    // 最大打开连接数 推荐值: 50-200 (说明: 根据并发量和数据库服务器配置调整，避免连接过多)
	Timeout         int    // 连接超时时间 推荐值: 5000-10000 毫秒 (5-10秒)
	Lifetime        int    // 连接最大生命周期 推荐值: 3600-7200 秒 (1-2小时，说明: 定期回收连接，避免长期使用导致的问题)
}
