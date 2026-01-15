package sqlite

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	// Import the SQLite driver.
	_ "modernc.org/sqlite"

	"github.com/wdmsyhh/simple-notes/internal/profile"
	"github.com/wdmsyhh/simple-notes/store"
)

// DB SQLite 数据库驱动实现
type DB struct {
	// db 数据库连接实例
	db      *sql.DB
	// profile 服务器配置
	profile *profile.Profile
}

// NewDB 打开一个数据库，由数据库驱动名称和驱动特定的数据源名称指定
// 参数：
//   profile - 服务器配置
// 返回：
//   store.Driver - 数据库驱动实例
//   error - 错误信息
func NewDB(profile *profile.Profile) (store.Driver, error) {
	// 确保在尝试打开数据库之前设置了 DSN
	if profile.DSN == "" {
		return nil, errors.New("dsn required")
	}

	// 确保数据库文件所在的目录存在
	dbDir := filepath.Dir(profile.DSN)
	if dbDir != "" && dbDir != "." {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, errors.Wrapf(err, "failed to create database directory: %s", dbDir)
		}
	}

	// 连接到数据库，使用一些合理的设置：
	// - 无共享缓存：已过时；WAL 日志模式是更好的解决方案
	// - 无外键约束：默认情况下已禁用，但明确设置是好的做法，可以防止 SQLite 升级时的意外
	// - 日志模式设置为 WAL：这是大多数应用程序推荐的日志模式，因为它可以防止锁定问题
	//
	// 注意：
	// - 使用 `modernc.org/sqlite` 驱动时，每个 pragma 必须前缀 `_pragma=`
	//
	// 参考：
	// - https://pkg.go.dev/modernc.org/sqlite#Driver.Open
	// - https://www.sqlite.org/sharedcache.html
	// - https://www.sqlite.org/pragma.html
	sqliteDB, err := sql.Open("sqlite", profile.DSN+"?_pragma=foreign_keys(0)&_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open db with dsn: %s", profile.DSN)
	}

	driver := DB{db: sqliteDB, profile: profile}

	return &driver, nil
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
	// 通过检查 notes 表是否存在来检查数据库是否已初始化
	var exists bool
	err := d.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='notes')").Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if database is initialized")
	}
	return exists, nil
}
