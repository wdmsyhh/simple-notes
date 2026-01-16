package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// UserRole 表示用户角色类型
type UserRole string

// 支持的用户角色
const (
	// RoleHost 系统管理员角色
	RoleHost UserRole = "HOST"
	// RoleAdmin 管理员角色
	RoleAdmin UserRole = "ADMIN"
	// RoleUser 普通用户角色
	RoleUser UserRole = "USER"
)

// User 表示系统中的用户
type User struct {
	// ID 用户ID
	ID uint `json:"id"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt 删除时间（软删除）
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	// Username 用户名
	Username string `json:"username"`
	// PasswordHash 密码哈希
	PasswordHash string `json:"password_hash"`
	// Nickname 昵称
	Nickname string `json:"nickname"`
	// Avatar 头像URL
	Avatar string `json:"avatar"`
	// Bio 个人简介
	Bio string `json:"bio"`
	// Role 用户角色
	Role UserRole `json:"role"`
}

// CreateUser 在数据库中创建新用户
func (s *Store) CreateUser(ctx context.Context, user *User) (*User, error) {
	// 检查是否已存在相同用户名的用户
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)"
	err := s.db.QueryRowContext(ctx, query, user.Username).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this username already exists")
	}

	// 插入新用户
	query = "INSERT INTO users (username, password_hash, nickname, avatar, bio, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"
	result, err := s.db.ExecContext(ctx, query, user.Username, user.PasswordHash, user.Nickname, user.Avatar, user.Bio, user.Role)
	if err != nil {
		return nil, err
	}

	// 获取最后插入的ID
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// 返回创建的用户
	return &User{
		ID:           uint(id),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Bio:          user.Bio,
		Role:         user.Role,
	}, nil
}

// GetUserByID 根据ID检索用户
func (s *Store) GetUserByID(ctx context.Context, id uint) (*User, error) {
	user := &User{}
	query := "SELECT id, created_at, updated_at, deleted_at, username, password_hash, nickname, avatar, bio, role FROM users WHERE id = ? AND deleted_at IS NULL"
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Username,
		&user.PasswordHash,
		&user.Nickname,
		&user.Avatar,
		&user.Bio,
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// GetUserByUsername 根据用户名检索用户
func (s *Store) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	query := "SELECT id, created_at, updated_at, deleted_at, username, password_hash, nickname, avatar, bio, role FROM users WHERE username = ? AND deleted_at IS NULL"
	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Username,
		&user.PasswordHash,
		&user.Nickname,
		&user.Avatar,
		&user.Bio,
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// ListUsers 检索所有用户
func (s *Store) ListUsers(ctx context.Context) ([]*User, error) {
	query := "SELECT id, created_at, updated_at, deleted_at, username, password_hash, nickname, avatar, bio, role FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
			&user.Username,
			&user.PasswordHash,
			&user.Nickname,
			&user.Avatar,
			&user.Bio,
			&user.Role,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUser 更新数据库中的现有用户
func (s *Store) UpdateUser(ctx context.Context, user *User) (*User, error) {
	query := "UPDATE users SET username = ?, password_hash = ?, nickname = ?, avatar = ?, bio = ?, role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL"
	_, err := s.db.ExecContext(ctx, query, user.Username, user.PasswordHash, user.Nickname, user.Avatar, user.Bio, user.Role, user.ID)
	if err != nil {
		return nil, err
	}

	// 返回更新后的用户
	return s.GetUserByID(ctx, user.ID)
}

// DeleteUser 从数据库中删除用户（软删除）
func (s *Store) DeleteUser(ctx context.Context, id uint) error {
	query := "UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL"
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// CountUsers 统计数据库中的用户数量
func (s *Store) CountUsers(ctx context.Context) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// IsUsernameAvailable 检查用户名是否可用
func (s *Store) IsUsernameAvailable(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = ? AND deleted_at IS NULL)"
	err := s.db.QueryRowContext(ctx, query, username).Scan(&exists)
	return !exists, err
}

