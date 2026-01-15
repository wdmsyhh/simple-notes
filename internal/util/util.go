package util

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// HasPrefixes 如果字符串 src 具有任何给定的前缀，则返回 true
// 参数：
//   src - 源字符串
//   prefixes - 前缀列表
// 返回：
//   bool - 是否具有任何前缀
func HasPrefixes(src string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(src, prefix) {
			return true
		}
	}
	return false
}

// ConvertStringToInt32 将字符串转换为 int32
// 参数：
//   src - 源字符串
// 返回：
//   int32 - 转换后的 int32 值
//   error - 错误信息
func ConvertStringToInt32(src string) (int32, error) {
	parsed, err := strconv.ParseInt(src, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(parsed), nil
}

// GenUUID 生成 UUID 字符串
// 返回：
//   string - UUID 字符串
func GenUUID() string {
	return uuid.New().String()
}

// letters 用于生成随机字符串的字符集
var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomString 返回长度为 n 的随机字符串
// 参数：
//   n - 字符串长度
// 返回：
//   string - 随机字符串
//   error - 错误信息
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

