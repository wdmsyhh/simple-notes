package store

import (
	"context"
	"database/sql"
)

// Driver 是存储驱动的接口
// 它包含所有存储数据库驱动应该实现的方法
type Driver interface {
	// GetDB 获取底层数据库连接
	GetDB() *sql.DB
	// Close 关闭数据库连接
	Close() error
	// IsInitialized 检查数据库是否已初始化
	IsInitialized(ctx context.Context) (bool, error)
}
