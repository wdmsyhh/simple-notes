package store

import (
	"context"
	"database/sql"
	"strconv"
)

const (
	// KeyLoginEnabled 是否允许登录/注册
	KeyLoginEnabled = "login_enabled"
)

// GetSetting 根据 key 获取设置值，不存在时返回空字符串
func (s *Store) GetSetting(ctx context.Context, key string) (string, error) {
	query := "SELECT value FROM system_settings WHERE key_name = ?"
	var value string
	err := s.db.QueryRowContext(ctx, query, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// SetSetting 设置 key 对应的值（跨数据库：先 UPDATE，无行时 INSERT）
func (s *Store) SetSetting(ctx context.Context, key, value string) error {
	result, err := s.db.ExecContext(ctx, "UPDATE system_settings SET value = ? WHERE key_name = ?", value, key)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		return nil
	}
	_, err = s.db.ExecContext(ctx, "INSERT INTO system_settings (key_name, value) VALUES (?, ?)", key, value)
	return err
}

// GetLoginEnabled 获取是否允许登录/注册，未设置时默认为 true
func (s *Store) GetLoginEnabled(ctx context.Context) (bool, error) {
	val, err := s.GetSetting(ctx, KeyLoginEnabled)
	if err != nil {
		return false, err
	}
	if val == "" {
		return true, nil
	}
	enabled, err := strconv.ParseBool(val)
	if err != nil {
		return true, nil
	}
	return enabled, nil
}

// SetLoginEnabled 设置是否允许登录/注册
func (s *Store) SetLoginEnabled(ctx context.Context, enabled bool) error {
	val := "false"
	if enabled {
		val = "true"
	}
	return s.SetSetting(ctx, KeyLoginEnabled, val)
}
