package util

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// HasPrefixes 如果字符串 src 具有任何给定的前缀，则返回 true
func HasPrefixes(src string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(src, prefix) {
			return true
		}
	}
	return false
}

// ConvertStringToInt32 将字符串转换为 int32
func ConvertStringToInt32(src string) (int32, error) {
	parsed, err := strconv.ParseInt(src, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(parsed), nil
}

// GenUUID 生成 UUID 字符串
func GenUUID() string {
	return uuid.New().String()
}

// letters 用于生成随机字符串的字符集
var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomString 返回长度为 n 的随机字符串
func RandomString(n int) (string, error) {
	var sb strings.Builder
	sb.Grow(n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		sb.WriteRune(letters[num.Int64()])
	}
	return sb.String(), nil
}
