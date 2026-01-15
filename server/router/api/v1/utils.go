package v1

import (
	"regexp"
	"strings"
	"time"
)

// generateSlug 从标题生成URL友好的slug
// 如果标题为空或仅包含非ASCII字符，则使用基于时间戳的slug
// 参数：
//
//	title - 标题字符串
//
// 返回：
//
//	string - 生成的slug
func generateSlug(title string) string {
	if title == "" {
		return generateTimestampSlug()
	}

	// 转换为小写
	slug := strings.ToLower(title)

	// 将空格和下划线替换为连字符
	slug = regexp.MustCompile(`[\s_]+`).ReplaceAllString(slug, "-")

	// 移除除连字符外的所有非单词字符
	slug = regexp.MustCompile(`[^\w\-]+`).ReplaceAllString(slug, "")

	// 将多个连字符替换为单个连字符
	slug = regexp.MustCompile(`\-+`).ReplaceAllString(slug, "-")

	// 移除前导和尾随的连字符
	slug = strings.Trim(slug, "-")

	// 如果处理后的slug为空，使用基于时间戳的slug
	if slug == "" {
		return generateTimestampSlug()
	}

	return slug
}

// generateTimestampSlug 基于时间戳生成slug
// 返回：
//
//	string - 基于时间戳的slug
func generateTimestampSlug() string {
	return "note-" + strings.ReplaceAll(time.Now().Format("20060102150405"), "-", "")
}
