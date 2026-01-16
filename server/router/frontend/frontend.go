package frontend

import (
	"context"
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/wdmsyhh/simple-notes/internal/profile"
	"github.com/wdmsyhh/simple-notes/internal/util"
	"github.com/wdmsyhh/simple-notes/store"
)

//go:embed dist/*
var embeddedFiles embed.FS

// FrontendService 前端静态文件服务
type FrontendService struct {
	// Profile 服务器配置
	Profile *profile.Profile
	// Store 数据存储实例
	Store *store.Store
}

// NewFrontendService 创建新的前端服务实例
func NewFrontendService(profile *profile.Profile, store *store.Store) *FrontendService {
	return &FrontendService{
		Profile: profile,
		Store:   store,
	}
}

// Serve 启动前端静态文件服务
func (*FrontendService) Serve(_ context.Context, e *echo.Echo) {
	skipper := func(c echo.Context) bool {
		// 跳过 API 路由
		if util.HasPrefixes(c.Path(), "/api", "/api.v1") {
			return true
		}
		// 跳过文件服务器路由
		if util.HasPrefixes(c.Path(), "/attachments") {
			return true
		}
		// 对于 index.html 和根路径，设置不缓存头部以防止浏览器缓存
		// 这可以防止敏感数据在登出后通过浏览器后退按钮访问
		if c.Path() == "/" || c.Path() == "/index.html" {
			c.Response().Header().Set(echo.HeaderCacheControl, "no-cache, no-store, must-revalidate")
			c.Response().Header().Set("Pragma", "no-cache")
			c.Response().Header().Set("Expires", "0")
			return false
		}
		// 为静态资源设置缓存控制头部
		// 由于 Vite 生成内容哈希的文件名（例如：index-BtVjejZf.js），
		// 我们可以积极缓存，但使用 immutable 来防止重新验证检查
		// 对于频繁重新部署的实例，使用较短的 max-age（1小时）以避免
		// 在重新部署后提供过时的资源
		c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=3600, immutable") // 1小时
		return false
	}

	// 使用 HTML5 fallback 为 SPA 行为提供主应用路由
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Filesystem: getFileSystem("dist"),
		HTML5:      true, // 启用回退到 index.html
		Skipper:    skipper,
	}))
}

// getFileSystem 获取文件系统
func getFileSystem(path string) http.FileSystem {
	fs, err := fs.Sub(embeddedFiles, path)
	if err != nil {
		panic(err)
	}
	return http.FS(fs)
}
