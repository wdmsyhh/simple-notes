package auth

import (
	"context"

	"github.com/pkg/errors"

	"github.com/wdmsyhh/simple-notes/internal/util"
	"github.com/wdmsyhh/simple-notes/store"
)

// Authenticator 提供共享的认证和授权逻辑
type Authenticator struct {
	// store 数据存储实例
	store *store.Store
	// secret JWT 密钥
	secret string
}

// NewAuthenticator 创建新的认证器实例
// 参数：
//
//	store - 数据存储实例
//	secret - JWT 密钥
//
// 返回：
//
//	*Authenticator - 认证器实例
func NewAuthenticator(store *store.Store, secret string) *Authenticator {
	return &Authenticator{
		store:  store,
		secret: secret,
	}
}

// AuthenticateByAccessTokenV2 验证短期访问令牌
// 参数：
//
//	accessToken - 访问令牌字符串
//
// 返回：
//
//	*UserClaims - 用户声明
//	error - 错误信息
func (a *Authenticator) AuthenticateByAccessTokenV2(accessToken string) (*UserClaims, error) {
	claims, err := ParseAccessTokenV2(accessToken, []byte(a.secret))
	if err != nil {
		return nil, errors.Wrap(err, "invalid access token")
	}

	userID, err := util.ConvertStringToInt32(claims.Subject)
	if err != nil {
		return nil, errors.Wrap(err, "invalid user ID in token")
	}

	return &UserClaims{
		UserID:   userID,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

// AuthResult 包含认证尝试的结果
type AuthResult struct {
	// Claims 用户声明，用于访问令牌 V2（无状态）
	Claims *UserClaims
	// AccessToken 访问令牌，如果通过 JWT 认证则非空
	AccessToken string
}

// Authenticate 尝试使用提供的凭据进行认证
// 参数：
//
//	ctx - 上下文
//	authHeader - 认证头部字符串
//
// 返回：
//
//	*AuthResult - 认证结果
func (a *Authenticator) Authenticate(ctx context.Context, authHeader string) *AuthResult {
	token := ExtractBearerToken(authHeader)

	// Try Access Token V2 (stateless)
	if token != "" {
		claims, err := a.AuthenticateByAccessTokenV2(token)
		if err == nil && claims != nil {
			return &AuthResult{
				Claims:      claims,
				AccessToken: token,
			}
		}
	}

	return nil
}
