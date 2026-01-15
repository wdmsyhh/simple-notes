package fileserver

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	storepb "github.com/wdmsyhh/simple-notes/proto/gen/store"
	"github.com/wdmsyhh/simple-notes/server/auth"
	"github.com/wdmsyhh/simple-notes/store"
)

// FileServerService 处理带有正确范围请求支持的 HTTP 文件服务
// 此服务绕过 gRPC-Gateway，使用原生 HTTP 服务通过 http.ServeContent()，
// 这对于 Safari 视频/音频播放是必需的
type FileServerService struct {
	// Store 数据存储实例
	Store         *store.Store
	// authenticator 认证器实例
	authenticator *auth.Authenticator
}

// NewFileServerService 创建新的文件服务器服务实例
// 参数：
//   store - 数据存储实例
//   secret - JWT 密钥
// 返回：
//   *FileServerService - 文件服务器服务实例
func NewFileServerService(store *store.Store, secret string) *FileServerService {
	return &FileServerService{
		Store:         store,
		authenticator: auth.NewAuthenticator(store, secret),
	}
}

// RegisterRoutes 注册 HTTP 文件服务路由
// 参数：
//   echoServer - Echo 服务器实例
func (s *FileServerService) RegisterRoutes(echoServer *echo.Echo) {
	fileGroup := echoServer.Group("/file")

	// 提供附件二进制文件服务
	fileGroup.GET("/attachments/:id/:filename", s.serveAttachmentFile)
}

// serveAttachmentFile 使用原生 HTTP 提供附件二进制内容服务
// 这正确处理了 Safari 视频/音频播放所需的范围请求
// 参数：
//   c - Echo 上下文
// 返回：
//   error - 错误信息
func (s *FileServerService) serveAttachmentFile(c echo.Context) error {
	ctx := c.Request().Context()
	idParam := c.Param("id")

	// 解析附件 ID
	var attachmentID int64
	if _, err := fmt.Sscanf(idParam, "%d", &attachmentID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid attachment ID")
	}

	// 从数据库获取附件
	attachment, err := s.Store.GetAttachment(ctx, attachmentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "attachment not found")
	}

	// 检查权限 - 如果附件属于某个笔记，验证笔记可见性
	if err := s.checkAttachmentPermission(ctx, c, attachment); err != nil {
		return err
	}

	// 获取二进制内容
	blob := attachment.Content
	if len(blob) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "attachment content not found")
	}

	// 确定内容类型
	contentType := attachment.Type
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if strings.HasPrefix(contentType, "text/") {
		contentType += "; charset=utf-8"
	}

	// 通过将潜在不安全的文件作为 octet-stream 提供来防止 XSS 攻击
	unsafeTypes := []string{
		"text/html",
		"text/javascript",
		"application/javascript",
		"application/x-javascript",
		"text/xml",
		"application/xml",
		"application/xhtml+xml",
		"image/svg+xml",
	}
	for _, unsafeType := range unsafeTypes {
		if strings.EqualFold(contentType, unsafeType) {
			contentType = "application/octet-stream"
			break
		}
	}

	// 设置通用头部
	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("Cache-Control", "public, max-age=3600")
	// 防止 MIME 类型嗅探，这可能导致 XSS
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	// 深度防御：防止嵌入到框架中并限制内容加载
	c.Response().Header().Set("X-Frame-Options", "DENY")
	c.Response().Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline';")

	// 对于非媒体文件强制下载以防止 XSS 执行
	if !strings.HasPrefix(contentType, "image/") &&
		!strings.HasPrefix(contentType, "video/") &&
		!strings.HasPrefix(contentType, "audio/") &&
		contentType != "application/pdf" {
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", attachment.Filename))
	}

	// 对于视频/音频：使用 http.ServeContent 自动支持范围请求
	// 这对于 Safari 至关重要，因为它需要范围请求支持
	if strings.HasPrefix(contentType, "video/") || strings.HasPrefix(contentType, "audio/") {
		// ServeContent 自动处理：
		// - 范围请求解析
		// - HTTP 206 部分内容响应
		// - Content-Range 头部
		// - Accept-Ranges: bytes 头部
		modTime := time.Unix(attachment.UpdatedAt, 0)
		if modTime.IsZero() {
			modTime = time.Now()
		}
		http.ServeContent(c.Response(), c.Request(), attachment.Filename, modTime, bytes.NewReader(blob))
		return nil
	}

	// 对于其他文件：简单的 blob 响应
	return c.Blob(http.StatusOK, contentType, blob)
}

// checkAttachmentPermission 验证用户是否有权限访问附件
// 参数：
//   ctx - 上下文
//   c - Echo 上下文
//   attachment - 附件对象
// 返回：
//   error - 错误信息
func (s *FileServerService) checkAttachmentPermission(ctx context.Context, c echo.Context, attachment *storepb.Attachment) error {
	// 如果附件未链接到笔记，检查用户是否是作者
	if attachment.NoteId == "" {
		// 对于未链接的附件，只有作者可以访问
		user, err := s.getCurrentUser(ctx, c)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get current user").SetInternal(err)
		}
		if user == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
		}

		// 检查当前用户是否是作者
		var authorID uint
		if _, err := fmt.Sscanf(attachment.AuthorId, "%d", &authorID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "invalid author ID format")
		}
		if user.ID != authorID {
			return echo.NewHTTPError(http.StatusForbidden, "forbidden access")
		}
		return nil
	}

	// 检查笔记可见性
	var noteID int64
	if _, err := fmt.Sscanf(attachment.NoteId, "notes/%d", &noteID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid note ID format")
	}

	note, err := s.Store.GetNote(ctx, noteID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "note not found")
	}

	// 公开笔记所有人都可以访问
	if note.Visibility == storepb.NoteVisibility_NOTE_VISIBILITY_PUBLIC {
		return nil
	}

	// 对于非公开笔记，检查认证
	user, err := s.getCurrentUser(ctx, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get current user").SetInternal(err)
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	// 私有笔记只能由创建者访问
	if note.Visibility == storepb.NoteVisibility_NOTE_VISIBILITY_PRIVATE {
		var authorID uint
		if _, err := fmt.Sscanf(note.AuthorId, "%d", &authorID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "invalid author ID format")
		}
		if user.ID != authorID {
			return echo.NewHTTPError(http.StatusForbidden, "forbidden access")
		}
	}

	return nil
}

// getCurrentUser 从 Echo 上下文检索当前已认证的用户
// 认证优先级：Bearer token（访问令牌 V2 或 PAT）> 刷新令牌 cookie
// 参数：
//   ctx - 上下文
//   c - Echo 上下文
// 返回：
//   *store.User - 用户对象
//   error - 错误信息
func (s *FileServerService) getCurrentUser(ctx context.Context, c echo.Context) (*store.User, error) {
	// 首先尝试 Bearer token 认证
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		token := auth.ExtractBearerToken(authHeader)
		if token != "" {
			// 尝试访问令牌 V2（无状态）
			result := s.authenticator.Authenticate(ctx, authHeader)
			if result != nil && result.Claims != nil {
				// 从声明中获取用户
				userID := uint(result.Claims.UserID)
				user, err := s.Store.GetUserByID(ctx, userID)
				if err == nil && user != nil {
					return user, nil
				}
			}
		}
	}

	// 未找到有效认证
	return nil, nil
}
