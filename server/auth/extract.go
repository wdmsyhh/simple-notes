package auth

import (
	"strings"
)

// ExtractBearerToken 从 Authorization 头部值中提取 JWT 令牌
// 预期格式："Bearer {token}"
// 如果未找到有效的 bearer 令牌，返回空字符串
// 参数：
//   authHeader - 认证头部字符串
// 返回：
//   string - 提取的令牌字符串
func ExtractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}

