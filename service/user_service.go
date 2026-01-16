package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/wdmsyhh/simple-notes/store"
)

// UserService 处理用户相关业务逻辑
type UserService struct {
	// store 数据存储实例
	store *store.Store
}

// NewUserService 创建新的用户服务实例
func NewUserService(store *store.Store) *UserService {
	return &UserService{
		store: store,
	}
}

// UserRegistrationRequest 表示用户注册请求
type UserRegistrationRequest struct {
	// Username 用户名
	Username string `json:"username"`
	// Email 邮箱
	Email string `json:"email"`
	// Password 密码
	Password string `json:"password"`
	// Nickname 昵称
	Nickname string `json:"nickname"`
	// Avatar 头像URL
	Avatar string `json:"avatar"`
	// Bio 个人简介
	Bio string `json:"bio"`
}

// UserLoginRequest 表示用户登录请求
type UserLoginRequest struct {
	// Username 用户名
	Username string `json:"username"`
	// Password 密码
	Password string `json:"password"`
}

// ValidateUserRegistrationRequest 验证用户注册请求
func ValidateUserRegistrationRequest(req *UserRegistrationRequest) error {
	// 验证用户名
	if strings.TrimSpace(req.Username) == "" {
		return errors.New("username is required")
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}
	// 用户名只能包含字母数字字符、下划线和连字符
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(req.Username) {
		return errors.New("username can only contain alphanumeric characters, underscores, and hyphens")
	}

	// 验证邮箱（可选）
	if strings.TrimSpace(req.Email) != "" {
		// 基本邮箱格式验证
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(req.Email) {
			return errors.New("invalid email format")
		}
	}

	// 验证密码
	if strings.TrimSpace(req.Password) == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	return nil
}

// HashPassword 使用 bcrypt 对密码进行哈希
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword 检查密码是否与哈希匹配
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// RegisterUser 注册新用户
func (s *UserService) RegisterUser(ctx context.Context, req *UserRegistrationRequest) (*store.User, error) {
	// 验证请求
	err := ValidateUserRegistrationRequest(req)
	if err != nil {
		return nil, err
	}

	// 检查用户注册是否启用
	// 目前我们假设它总是启用的
	// 稍后我们将添加设置来控制这一点

	// 对密码进行哈希
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 确定新用户的角色
	// 如果还没有用户，第一个用户获得 HOST 角色
	userCount, err := s.store.CountUsers(ctx)
	if err != nil {
		return nil, err
	}

	role := store.RoleUser
	if userCount == 0 {
		role = store.RoleHost
	}

	// 创建用户
	user := &store.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Nickname:     req.Nickname,
		Avatar:       req.Avatar,
		Bio:          req.Bio,
		Role:         role,
	}

	return s.store.CreateUser(ctx, user)
}

// LoginUser 认证用户
func (s *UserService) LoginUser(ctx context.Context, req *UserLoginRequest) (*store.User, error) {
	// 验证请求
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, errors.New("username and password are required")
	}

	// 根据用户名查找用户
	user, err := s.store.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	// 检查密码
	if !CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

// GetUserByID 根据ID检索用户
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*store.User, error) {
	return s.store.GetUserByID(ctx, id)
}

// GetUserByUsername 根据用户名检索用户
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*store.User, error) {
	return s.store.GetUserByUsername(ctx, username)
}

// GetUserByEmail 根据邮箱检索用户
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*store.User, error) {
	return s.store.GetUserByEmail(ctx, email)
}

// ListUsers 检索所有用户
func (s *UserService) ListUsers(ctx context.Context) ([]*store.User, error) {
	return s.store.ListUsers(ctx)
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, user *store.User) (*store.User, error) {
	return s.store.UpdateUser(ctx, user)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	return s.store.DeleteUser(ctx, id)
}

// IsSuperUser 检查用户是否是超级用户（HOST 或 ADMIN）
func IsSuperUser(user *store.User) bool {
	return user != nil && (user.Role == store.RoleHost || user.Role == store.RoleAdmin)
}

// IsHost 检查用户是否是 HOST
func IsHost(user *store.User) bool {
	return user != nil && user.Role == store.RoleHost
}

// IsAdmin 检查用户是否是 ADMIN
func IsAdmin(user *store.User) bool {
	return user != nil && user.Role == store.RoleAdmin
}

// IsRegularUser 检查用户是否是普通用户
func IsRegularUser(user *store.User) bool {
	return user != nil && user.Role == store.RoleUser
}
