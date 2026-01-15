package auth

import (
	"context"
)

// ContextKey 上下文值的键类型
type ContextKey int

const (
	// UserIDContextKey 存储已认证用户的ID
	UserIDContextKey ContextKey = iota

	// AccessTokenContextKey 存储用于基于令牌认证的 JWT 令牌
	AccessTokenContextKey

	// UserClaimsContextKey 存储来自访问令牌的声明
	UserClaimsContextKey
)

// GetUserID 从上下文中检索已认证用户的ID
// 参数：
//   ctx - 上下文
// 返回：
//   int32 - 用户ID，如果未找到则返回0
func GetUserID(ctx context.Context) int32 {
	if v, ok := ctx.Value(UserIDContextKey).(int32); ok {
		return v
	}
	return 0
}

// GetUserClaims 从上下文中检索用户声明
// 参数：
//   ctx - 上下文
// 返回：
//   *UserClaims - 用户声明，如果未找到则返回nil
func GetUserClaims(ctx context.Context) *UserClaims {
	if v, ok := ctx.Value(UserClaimsContextKey).(*UserClaims); ok {
		return v
	}
	return nil
}

// SetUserClaimsInContext 在上下文中设置用户声明
// 参数：
//   ctx - 上下文
//   claims - 用户声明
// 返回：
//   context.Context - 更新后的上下文
func SetUserClaimsInContext(ctx context.Context, claims *UserClaims) context.Context {
	ctx = context.WithValue(ctx, UserClaimsContextKey, claims)
	ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
	return ctx
}

