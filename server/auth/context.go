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
func GetUserID(ctx context.Context) int32 {
	if v, ok := ctx.Value(UserIDContextKey).(int32); ok {
		return v
	}
	return 0
}

// GetUserClaims 从上下文中检索用户声明
func GetUserClaims(ctx context.Context) *UserClaims {
	if v, ok := ctx.Value(UserClaimsContextKey).(*UserClaims); ok {
		return v
	}
	return nil
}

// SetUserClaimsInContext 在上下文中设置用户声明
func SetUserClaimsInContext(ctx context.Context, claims *UserClaims) context.Context {
	ctx = context.WithValue(ctx, UserClaimsContextKey, claims)
	ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
	return ctx
}
