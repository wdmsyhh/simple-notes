package mysql

import (
	"context"
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"github.com/wdmsyhh/simple-notes/internal/profile"
	"github.com/wdmsyhh/simple-notes/store"
)

// DB MySQL 数据库驱动实现
type DB struct {
	// db 数据库连接实例
	db *sql.DB
	// profile 服务器配置
	profile *profile.Profile
	// config MySQL配置
	config *mysql.Config
}

// NewDB 创建新的MySQL数据库驱动实例
func NewDB(profile *profile.Profile) (store.Driver, error) {
	// 打开 MySQL 连接，带参数
	// multiStatements=true 是迁移所必需的
	// 参考: https://github.com/go-sql-driver/mysql#multistatements
	dsn, err := mergeDSN(profile.DSN)
	if err != nil {
		return nil, err
	}

	driver := DB{profile: profile}
	driver.config, err = mysql.ParseDSN(dsn)
	if err != nil {
		return nil, errors.New("Parse DSN error")
	}

	driver.db, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open db: %s", profile.DSN)
	}

	return &driver, nil
}

// GetDB 获取底层数据库连接
func (d *DB) GetDB() *sql.DB {
	return d.db
}

// Close 关闭数据库连接
func (d *DB) Close() error {
	return d.db.Close()
}

// IsInitialized 检查数据库是否已初始化
func (d *DB) IsInitialized(ctx context.Context) (bool, error) {
	var exists bool
	err := d.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'notes' AND TABLE_TYPE = 'BASE TABLE')").Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if database is initialized")
	}
	return exists, nil
}

// mergeDSN 合并DSN配置，添加必要的参数
func mergeDSN(baseDSN string) (string, error) {
	config, err := mysql.ParseDSN(baseDSN)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse DSN: %s", baseDSN)
	}

	config.MultiStatements = true
	return config.FormatDSN(), nil
}
