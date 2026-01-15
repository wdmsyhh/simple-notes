package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"

	"github.com/wdmsyhh/simple-notes/internal/util"
)

const (
	// Issuer JWT 令牌中的发行者声明
	Issuer = "simple-notes"

	// KeyID JWT 头部中使用的密钥标识符
	KeyID = "v1"

	// AccessTokenAudienceName JWT 访问令牌的受众声明
	AccessTokenAudienceName = "user.access-token"

	// AccessTokenDuration 访问令牌的生命周期（1天）
	AccessTokenDuration = 24 * time.Hour
)

// AccessTokenClaims 包含短期访问令牌的声明
// 这些令牌仅通过签名验证（无状态）
type AccessTokenClaims struct {
	// Type 令牌类型，值为 "access"
	Type     string `json:"type"`
	// Role 用户角色
	Role     string `json:"role"`
	// Username 用于显示的用户名
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// UserClaims 表示来自访问令牌的已认证用户信息
type UserClaims struct {
	// UserID 用户ID
	UserID   int32
	// Username 用户名
	Username string
	// Role 用户角色
	Role     string
}

// GenerateAccessTokenV2 生成带有用户声明的短期访问令牌
// 参数：
//   userID - 用户ID
//   username - 用户名
//   role - 用户角色
//   secret - JWT 密钥
// 返回：
//   string - 令牌字符串
//   time.Time - 过期时间
//   error - 错误信息
func GenerateAccessTokenV2(userID int32, username, role string, secret []byte) (string, time.Time, error) {
	expiresAt := time.Now().Add(AccessTokenDuration)

	claims := &AccessTokenClaims{
		Type:     "access",
		Role:     role,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Audience:  jwt.ClaimStrings{AccessTokenAudienceName},
			Subject:   fmt.Sprint(userID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = KeyID

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// verifyJWTKeyFunc 返回一个验证签名方法和密钥ID的 jwt.Keyfunc
// 参数：
//   secret - JWT 密钥
// 返回：
//   jwt.Keyfunc - JWT 密钥函数
func verifyJWTKeyFunc(secret []byte) jwt.Keyfunc {
	return func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, errors.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		kid, ok := t.Header["kid"].(string)
		if !ok || kid != KeyID {
			return nil, errors.Errorf("unexpected kid: %v", t.Header["kid"])
		}
		return secret, nil
	}
}

// ParseAccessTokenV2 解析并验证短期访问令牌
// 参数：
//   tokenString - 令牌字符串
//   secret - JWT 密钥
// 返回：
//   *AccessTokenClaims - 访问令牌声明
//   error - 错误信息
func ParseAccessTokenV2(tokenString string, secret []byte) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, verifyJWTKeyFunc(secret),
		jwt.WithIssuer(Issuer),
		jwt.WithAudience(AccessTokenAudienceName),
	)
	if err != nil {
		return nil, err
	}
	if claims.Type != "access" {
		return nil, errors.New("invalid token type: expected access token")
	}
	return claims, nil
}

// RandomString 返回长度为 n 的随机字符串
// 参数：
//   n - 字符串长度
// 返回：
//   string - 随机字符串
//   error - 错误信息
func RandomString(n int) (string, error) {
	return util.RandomString(n)
}

