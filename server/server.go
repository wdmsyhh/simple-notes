package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/wdmsyhh/simple-notes/internal/profile"
	apiv1 "github.com/wdmsyhh/simple-notes/server/router/api/v1"
	"github.com/wdmsyhh/simple-notes/server/router/fileserver"
	"github.com/wdmsyhh/simple-notes/server/router/frontend"
	"github.com/wdmsyhh/simple-notes/store"
)

// Server 表示应用服务器
// 封装了Echo框架实例、数据存储和端口配置

type Server struct {
	// Store - 数据存储实例，用于数据库操作
	Store *store.Store
	// Profile - 服务器配置
	Profile *profile.Profile
	// Port - 服务器监听端口
	Port int
	// echoServer - Echo框架实例，处理HTTP请求
	echoServer *echo.Echo
}

// NewServer 创建一个新的服务器实例
func NewServer(store *store.Store, profile *profile.Profile, port int) *Server {
	echoServer := echo.New()
	echoServer.Debug = true
	echoServer.HideBanner = true
	echoServer.HidePort = true
	// 使用恢复中间件，处理panic
	echoServer.Use(middleware.Recover())
	// 设置请求体大小限制为 32MB（与附件上传限制一致）
	echoServer.Use(middleware.BodyLimit("32M"))

	return &Server{
		Store:      store,
		Profile:    profile,
		Port:       port,
		echoServer: echoServer,
	}
}

// SetupRoutes 配置所有路由和服务
func (s *Server) SetupRoutes(ctx context.Context) error {
	// 注册健康检查端点
	s.echoServer.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "Service ready.")
	})

	// 生成或使用默认 secret（生产环境应该从环境变量或配置文件读取）
	secret := os.Getenv("SECRET")
	if secret == "" {
		secret = "simple-notes-secret-key-change-in-production"
	}

	// Serve frontend static files.
	frontend.NewFrontendService(s.Profile, s.Store).Serve(ctx, s.echoServer)

	// Register HTTP file server routes BEFORE gRPC-Gateway to ensure proper range request handling for Safari.
	fileServerService := fileserver.NewFileServerService(s.Store, secret)
	fileServerService.RegisterRoutes(s.echoServer)

	// 注册API v1服务
	apiV1Service := apiv1.NewAPIV1Service(s.Store, secret)
	if err := apiV1Service.RegisterGateway(ctx, s.echoServer); err != nil {
		return fmt.Errorf("failed to register API v1 gateway: %w", err)
	}

	log.Println("Routes configured")
	return nil
}

// Start 启动服务器
func (s *Server) Start() error {
	address := fmt.Sprintf(":%d", s.Port)
	log.Printf("Server starting on %s", address)
	return s.echoServer.Start(address)
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Server shutting down")
	return s.echoServer.Shutdown(ctx)
}
