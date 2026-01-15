package postgres

import (
	"context"
	"database/sql"
	"log"

	// Import the PostgreSQL driver.
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/wdmsyhh/simple-notes/internal/profile"
	"github.com/wdmsyhh/simple-notes/store"
)

// DB PostgreSQL 数据库驱动实现
type DB struct {
	// db 数据库连接实例
	db      *sql.DB
	// profile 服务器配置
	profile *profile.Profile
}

// NewDB 创建新的PostgreSQL数据库驱动实例
// 参数：
//   profile - 服务器配置
// 返回：
//   store.Driver - 数据库驱动实例
//   error - 错误信息
func NewDB(profile *profile.Profile) (store.Driver, error) {
	if profile == nil {
		return nil, errors.New("profile is nil")
	}

	// 打开 PostgreSQL 连接
	db, err := sql.Open("postgres", profile.DSN)
	if err != nil {
		log.Printf("Failed to open database: %s", err)
		return nil, errors.Wrapf(err, "failed to open database: %s", profile.DSN)
	}

	var driver store.Driver = &DB{
		db:      db,
		profile: profile,
	}

	// 返回 DB 结构体
	return driver, nil
}

// GetDB 获取底层数据库连接
// 返回：
//   *sql.DB - 数据库连接实例
func (d *DB) GetDB() *sql.DB {
	return d.db
}

// Close 关闭数据库连接
// 返回：
//   error - 错误信息
func (d *DB) Close() error {
	return d.db.Close()
}

// IsInitialized 检查数据库是否已初始化
// 参数：
//   ctx - 上下文
// 返回：
//   bool - 是否已初始化
//   error - 错误信息
func (d *DB) IsInitialized(ctx context.Context) (bool, error) {
	var exists bool
	err := d.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_catalog = current_database() AND table_name = 'notes' AND table_type = 'BASE TABLE')").Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if database is initialized")
	}
	return exists, nil
}
