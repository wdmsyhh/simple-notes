package db

import (
	"github.com/pkg/errors"

	"github.com/wdmsyhh/simple-notes/internal/profile"
	"github.com/wdmsyhh/simple-notes/store"
	"github.com/wdmsyhh/simple-notes/store/db/mysql"
	"github.com/wdmsyhh/simple-notes/store/db/postgres"
	"github.com/wdmsyhh/simple-notes/store/db/sqlite"
)

// NewDBDriver 根据 profile 创建新的数据库驱动
// 参数：
//   profile - 服务器配置
// 返回：
//   store.Driver - 数据库驱动实例
//   error - 错误信息
func NewDBDriver(profile *profile.Profile) (store.Driver, error) {
	var driver store.Driver
	var err error

	switch profile.Driver {
	case "sqlite":
		driver, err = sqlite.NewDB(profile)
	case "mysql":
		driver, err = mysql.NewDB(profile)
	case "postgres":
		driver, err = postgres.NewDB(profile)
	default:
		return nil, errors.New("unknown db driver")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to create db driver")
	}
	return driver, nil
}
