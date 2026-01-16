package auth

import (
	"strings"
)

// ExtractBearerToken 从 Authorization 头部值中提取 JWT 令牌
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
